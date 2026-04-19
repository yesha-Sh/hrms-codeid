package assignments

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	internalauth "hrms/backend/internal/auth"
	"hrms/backend/internal/httpx"
	"hrms/backend/internal/middleware"
	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/audit"
	"hrms/backend/internal/modules/shared"
)

type Service struct {
	db    *gorm.DB
	audit audit.Service
}

type upsertRequest struct {
	EmployeeID            string   `json:"employee_id"`
	JobID                 string   `json:"job_id"`
	DepartmentID          string   `json:"department_id"`
	EstimatedHoursPerWeek *float64 `json:"estimated_hours_per_week"`
	StartDate             string   `json:"start_date"`
	EndDate               *string  `json:"end_date"`
	Notes                 string   `json:"notes"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/employee-role-assignments", func(r chi.Router) {
		r.Get("/", service.list)
		r.Post("/", service.create)
		r.Get("/{id}", service.get)
		r.Put("/{id}", service.update)
		r.Delete("/{id}", service.delete)
	})
}

func (s Service) list(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	pagination := shared.ParsePagination(r, "created_at")
	query := s.baseQuery()

	if search := r.URL.Query().Get("search"); search != "" {
		term := "%" + search + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR jobs.title ILIKE ? OR departments.name ILIKE ?", term, term, term, term)
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		query = query.Where("employee_role_assignments.employee_id = ?", employeeID)
	}
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("employee_role_assignments.department_id = ?", departmentID)
	}
	if jobID := r.URL.Query().Get("job_id"); jobID != "" {
		query = query.Where("employee_role_assignments.job_id = ?", jobID)
	}

	scoped, err := s.applyScope(query, claims)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var total int64
	if err := scoped.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count assignments", nil)
		return
	}

	var items []models.EmployeeRoleAssignment
	if err := scoped.Order("employee_role_assignments." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&items).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list assignments", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": toListResponse(items), "meta": map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total}})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid assignment id", nil)
		return
	}

	item, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "assignment not found", nil)
		return
	}
	if err := s.ensureAccess(item, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	httpx.JSON(w, http.StatusOK, toResponse(item))
}

func (s Service) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "only admins can manage secondary assignments", nil)
		return
	}

	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	item, err := s.toModel(req, claims)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := s.validateAssignmentWindow(item, nil); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := s.db.Create(&item).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create assignment", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "employee_role_assignment", &item.ID, map[string]any{"employee_id": item.EmployeeID})

	fresh, _ := s.findByID(item.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "only admins can manage secondary assignments", nil)
		return
	}

	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid assignment id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "assignment not found", nil)
		return
	}
	if err := s.ensureAccess(current, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updated, err := s.toModel(req, claims)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	updated.ID = current.ID
	updated.OrganizationID = current.OrganizationID
	if err := s.validateAssignmentWindow(updated, &current.ID); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := s.db.Model(&current).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update assignment", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "employee_role_assignment", &current.ID, map[string]any{"employee_id": updated.EmployeeID})

	fresh, _ := s.findByID(current.ID)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "only admins can manage secondary assignments", nil)
		return
	}

	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid assignment id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "assignment not found", nil)
		return
	}
	if err := s.ensureAccess(current, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	if err := s.db.Delete(&models.EmployeeRoleAssignment{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete assignment", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "employee_role_assignment", &id, nil)
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "assignment deleted"})
}

func (s Service) baseQuery() *gorm.DB {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return s.db.Model(&models.EmployeeRoleAssignment{}).Where("1 = 0")
	}
	return s.db.Model(&models.EmployeeRoleAssignment{}).
		Joins("JOIN employees ON employees.id = employee_role_assignments.employee_id").
		Joins("JOIN jobs ON jobs.id = employee_role_assignments.job_id").
		Joins("JOIN departments ON departments.id = employee_role_assignments.department_id").
		Where("employee_role_assignments.organization_id = ?", organizationID).
		Preload("Employee").
		Preload("Employee.Department").
		Preload("Job.JobLevel").
		Preload("Department")
}

func (s Service) applyScope(query *gorm.DB, claims *internalauth.AccessClaims) (*gorm.DB, error) {
	if claims == nil {
		return nil, errors.New("missing auth claims")
	}
	switch claims.Role {
	case "admin":
		return query, nil
	case "manager":
		managerID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || managerID == nil {
			return nil, errors.New("manager employee context is required")
		}
		departmentIDs, err := shared.AccessibleDepartmentIDs(s.db, *managerID)
		if err != nil {
			return nil, err
		}
		return query.Where("employee_role_assignments.employee_id = ? OR employees.manager_employee_id = ? OR employees.department_id IN ?", *managerID, *managerID, departmentIDs), nil
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return nil, errors.New("employee context is required")
		}
		return query.Where("employee_role_assignments.employee_id = ?", *employeeID), nil
	default:
		return nil, errors.New("unsupported role")
	}
}

func (s Service) ensureAccess(item models.EmployeeRoleAssignment, claims *internalauth.AccessClaims) error {
	if claims == nil {
		return errors.New("missing auth claims")
	}
	switch claims.Role {
	case "admin":
		return nil
	case "manager":
		managerID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || managerID == nil {
			return errors.New("manager employee context is required")
		}
		departmentIDs, err := shared.AccessibleDepartmentIDs(s.db, *managerID)
		if err != nil {
			return err
		}
		if item.EmployeeID == *managerID || shared.EmployeeWithinManagerScope(item.Employee, *managerID, departmentIDs) {
			return nil
		}
		return errors.New("assignment is outside manager scope")
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return errors.New("employee context is required")
		}
		if item.EmployeeID != *employeeID {
			return errors.New("assignment is outside user scope")
		}
		return nil
	default:
		return errors.New("unsupported role")
	}
}

func (s Service) findByID(id uuid.UUID) (models.EmployeeRoleAssignment, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	var item models.EmployeeRoleAssignment
	err = s.baseQuery().First(&item, "employee_role_assignments.id = ? AND employee_role_assignments.organization_id = ?", id, organizationID).Error
	return item, err
}

func (s Service) toModel(req upsertRequest, claims *internalauth.AccessClaims) (models.EmployeeRoleAssignment, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	employeeID, err := shared.ParseUUID(req.EmployeeID)
	if err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	jobID, err := shared.ParseUUID(req.JobID)
	if err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	departmentID, err := shared.ParseUUID(req.DepartmentID)
	if err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return models.EmployeeRoleAssignment{}, err
		}
		endDate = &parsed
	}
	hours := 0.0
	if req.EstimatedHoursPerWeek != nil {
		hours = *req.EstimatedHoursPerWeek
	}

	item := models.EmployeeRoleAssignment{OrganizationID: organizationID, EmployeeID: employeeID, JobID: jobID, DepartmentID: departmentID, EstimatedHoursPerWeek: hours, StartDate: startDate, EndDate: endDate, Notes: req.Notes}
	if err := s.ensureWritable(item, claims); err != nil {
		return models.EmployeeRoleAssignment{}, err
	}
	return item, nil
}

func (s Service) ensureWritable(item models.EmployeeRoleAssignment, claims *internalauth.AccessClaims) error {
	if claims == nil {
		return errors.New("missing auth claims")
	}
	if claims.Role != "admin" {
		return errors.New("only admins can manage secondary assignments")
	}

	if item.EstimatedHoursPerWeek <= 0 {
		return errors.New("estimated_hours_per_week must be greater than zero")
	}
	if item.EstimatedHoursPerWeek > 40 {
		return errors.New("estimated_hours_per_week cannot exceed 40 hours")
	}
	if item.EndDate != nil && item.EndDate.Before(item.StartDate) {
		return errors.New("end_date cannot be earlier than start_date")
	}

	return nil
}

func (s Service) validateAssignmentWindow(item models.EmployeeRoleAssignment, excludeID *uuid.UUID) error {
	query := s.db.Model(&models.EmployeeRoleAssignment{}).
		Where("organization_id = ? AND employee_id = ?", item.OrganizationID, item.EmployeeID)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var existing []models.EmployeeRoleAssignment
	if err := query.Find(&existing).Error; err != nil {
		return err
	}

	totalHours := item.EstimatedHoursPerWeek
	for _, current := range existing {
		if !windowsOverlap(item.StartDate, item.EndDate, current.StartDate, current.EndDate) {
			continue
		}
		if current.JobID == item.JobID && current.DepartmentID == item.DepartmentID {
			return errors.New("an overlapping assignment already exists for this employee, job, and department")
		}
		totalHours += current.EstimatedHoursPerWeek
	}

	if totalHours > 40 {
		return errors.New("overlapping assignment workload cannot exceed 40 hours per week")
	}

	return nil
}

func windowsOverlap(startA time.Time, endA *time.Time, startB time.Time, endB *time.Time) bool {
	finishA := endOfWindow(endA)
	finishB := endOfWindow(endB)
	return !finishA.Before(startB) && !finishB.Before(startA)
}

func endOfWindow(value *time.Time) time.Time {
	if value == nil {
		return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	}
	return value.UTC().Truncate(24 * time.Hour)
}

func toListResponse(items []models.EmployeeRoleAssignment) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.EmployeeRoleAssignment) map[string]any {
	assignmentStatus := shared.AssignmentStatus(item.StartDate, item.EndDate, time.Now())
	response := map[string]any{
		"id":                       item.ID,
		"organization_id":          item.OrganizationID,
		"employee_id":              item.EmployeeID,
		"employee_name":            shared.FullName(item.Employee.FirstName, item.Employee.LastName),
		"job_id":                   item.JobID,
		"job_title":                item.Job.Title,
		"job_level_name":           item.Job.JobLevel.Name,
		"department_id":            item.DepartmentID,
		"department_name":          item.Department.Name,
		"estimated_hours_per_week": item.EstimatedHoursPerWeek,
		"start_date":               item.StartDate.Format("2006-01-02"),
		"assignment_status":        assignmentStatus,
		"notes":                    item.Notes,
		"created_at":               item.CreatedAt,
		"updated_at":               item.UpdatedAt,
	}
	if item.EndDate != nil {
		response["end_date"] = item.EndDate.Format("2006-01-02")
	}
	return response
}
