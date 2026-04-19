package employees

import (
	"errors"
	"net/http"
	"strings"
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
	EmployeeCode       string   `json:"employee_code"`
	FirstName          string   `json:"first_name"`
	LastName           string   `json:"last_name"`
	Email              string   `json:"email"`
	PhoneNumber        string   `json:"phone_number"`
	HireDate           string   `json:"hire_date"`
	Salary             *float64 `json:"salary"`
	EmployeeTypeID     string   `json:"employee_type_id"`
	EmploymentStatusID string   `json:"employment_status_id"`
	DepartmentID       string   `json:"department_id"`
	JobID              string   `json:"job_id"`
	LocationID         string   `json:"location_id"`
	WorkMode           string   `json:"work_mode"`
	ManagementScope    string   `json:"management_scope"`
	ManagerEmployeeID  *string  `json:"manager_employee_id"`
	UserID             *string  `json:"user_id"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/employees", func(r chi.Router) {
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
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR employees.email ILIKE ? OR employees.employee_code ILIKE ?", term, term, term, term)
	}
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("employees.department_id = ?", departmentID)
	}
	if locationID := r.URL.Query().Get("location_id"); locationID != "" {
		query = query.Where("employees.location_id = ?", locationID)
	}
	if employeeTypeID := r.URL.Query().Get("employee_type_id"); employeeTypeID != "" {
		query = query.Where("employees.employee_type_id = ?", employeeTypeID)
	}
	if employmentStatusID := r.URL.Query().Get("employment_status_id"); employmentStatusID != "" {
		query = query.Where("employees.employment_status_id = ?", employmentStatusID)
	}
	if managementScope := r.URL.Query().Get("management_scope"); managementScope != "" {
		query = query.Where("employees.management_scope = ?", strings.ToLower(strings.TrimSpace(managementScope)))
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Joins("JOIN employment_statuses ON employment_statuses.id = employees.employment_status_id").Where("employment_statuses.code = ?", strings.ToLower(status))
	}

	scoped, err := s.applyScope(query, claims)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var total int64
	if err := scoped.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count employees", nil)
		return
	}

	var employees []models.Employee
	if err := scoped.Order("employees." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&employees).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list employees", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"items": toListResponse(employees),
		"meta":  map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total},
	})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid employee id", nil)
		return
	}

	employee, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "employee not found", nil)
		return
	}

	if err := s.ensureAccess(employee, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	httpx.JSON(w, http.StatusOK, toResponse(employee))
}

func (s Service) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "only admins can create employee records", nil)
		return
	}
	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	employee, err := s.toModel(req, claims)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := s.db.Create(&employee).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create employee", nil)
		return
	}

	if employee.UserID != nil {
		_ = s.db.Model(&models.User{}).Where("id = ?", *employee.UserID).Update("employee_id", employee.ID).Error
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "employee", &employee.ID, map[string]any{"email": employee.Email})

	fresh, _ := s.findByID(employee.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid employee id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "employee not found", nil)
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

	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "only admins can edit employee records", nil)
		return
	}

	updated, err := s.toModel(req, claims)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	updated.ID = current.ID
	updated.OrganizationID = current.OrganizationID

	if err := s.db.Model(&current).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update employee", nil)
		return
	}

	if updated.UserID != nil {
		_ = s.db.Model(&models.User{}).Where("id = ?", *updated.UserID).Update("employee_id", current.ID).Error
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "employee", &current.ID, map[string]any{"email": updated.Email})

	fresh, _ := s.findByID(current.ID)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "only admins can delete employee records", nil)
		return
	}
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid employee id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "employee not found", nil)
		return
	}
	if err := s.ensureAccess(current, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	if err := s.db.Delete(&models.Employee{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete employee", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "employee", &id, nil)
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "employee deleted"})
}

func (s Service) baseQuery() *gorm.DB {
	organizationID, _ := shared.MustOrganizationID(s.db)
	return shared.PreloadEmployeeDetails(s.db.Model(&models.Employee{})).Where("employees.organization_id = ?", organizationID)
}

func (s Service) applyScope(query *gorm.DB, claims *internalauth.AccessClaims) (*gorm.DB, error) {
	if claims == nil {
		return nil, errors.New("missing auth claims")
	}
	switch claims.Role {
	case "admin":
		return query, nil
	case "manager":
		managerID, departmentIDs, err := s.managerScope(claims)
		if err != nil {
			return nil, err
		}
		return query.Where("employees.manager_employee_id = ? OR employees.department_id IN ?", managerID, departmentIDs), nil
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return nil, errors.New("employee context is required")
		}
		return query.Where("employees.id = ?", *employeeID), nil
	default:
		return nil, errors.New("unsupported role")
	}
}

func (s Service) ensureAccess(employee models.Employee, claims *internalauth.AccessClaims) error {
	switch claims.Role {
	case "admin":
		return nil
	case "manager":
		managerID, departmentIDs, err := s.managerScope(claims)
		if err != nil {
			return err
		}
		if !shared.EmployeeWithinManagerScope(employee, managerID, departmentIDs) {
			return errors.New("employee is outside manager scope")
		}
		return nil
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return errors.New("employee context is required")
		}
		if employee.ID != *employeeID {
			return errors.New("employee is outside user scope")
		}
		return nil
	default:
		return errors.New("unsupported role")
	}
}

func (s Service) findByID(id uuid.UUID) (models.Employee, error) {
	var employee models.Employee
	err := s.baseQuery().First(&employee, "employees.id = ?", id).Error
	return employee, err
}

func (s Service) toModel(req upsertRequest, claims *internalauth.AccessClaims) (models.Employee, error) {
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return models.Employee{}, err
	}

	employeeTypeID, err := uuid.Parse(req.EmployeeTypeID)
	if err != nil {
		return models.Employee{}, err
	}
	employmentStatusID, err := uuid.Parse(req.EmploymentStatusID)
	if err != nil {
		return models.Employee{}, err
	}
	departmentID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		return models.Employee{}, err
	}
	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		return models.Employee{}, err
	}
	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		return models.Employee{}, err
	}
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Employee{}, err
	}

	employee := models.Employee{
		OrganizationID:     organizationID,
		EmployeeCode:       req.EmployeeCode,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Email:              strings.ToLower(req.Email),
		PhoneNumber:        req.PhoneNumber,
		HireDate:           hireDate,
		Salary:             req.Salary,
		EmployeeTypeID:     employeeTypeID,
		EmploymentStatusID: employmentStatusID,
		DepartmentID:       departmentID,
		JobID:              jobID,
		LocationID:         locationID,
		WorkMode:           normalizeWorkMode(req.WorkMode),
		ManagementScope:    normalizeManagementScope(req.ManagementScope),
	}

	if req.ManagerEmployeeID != nil && *req.ManagerEmployeeID != "" {
		managerID, err := uuid.Parse(*req.ManagerEmployeeID)
		if err != nil {
			return models.Employee{}, err
		}
		employee.ManagerEmployeeID = &managerID
	}

	if req.UserID != nil && *req.UserID != "" {
		userID, err := uuid.Parse(*req.UserID)
		if err != nil {
			return models.Employee{}, err
		}
		employee.UserID = &userID
	}

	return employee, nil
}

func (s Service) managerScope(claims *internalauth.AccessClaims) (uuid.UUID, []uuid.UUID, error) {
	managerID, err := shared.CurrentEmployeeUUID(claims)
	if err != nil || managerID == nil {
		return uuid.Nil, nil, errors.New("manager employee context is required")
	}
	context, err := shared.LoadManagerContext(s.db, *managerID)
	if err != nil {
		return uuid.Nil, nil, err
	}
	return *managerID, context.ManagedDepartmentIDs, nil
}

func toListResponse(items []models.Employee) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.Employee) map[string]any {
	response := map[string]any{
		"id":                     item.ID,
		"organization_id":        item.OrganizationID,
		"employee_code":          item.EmployeeCode,
		"first_name":             item.FirstName,
		"last_name":              item.LastName,
		"full_name":              shared.FullName(item.FirstName, item.LastName),
		"email":                  item.Email,
		"phone_number":           item.PhoneNumber,
		"hire_date":              item.HireDate.Format("2006-01-02"),
		"salary":                 item.Salary,
		"employee_type_id":       item.EmployeeTypeID,
		"employee_type_name":     item.EmployeeType.Name,
		"employment_status_id":   item.EmploymentStatusID,
		"employment_status_name": item.EmploymentStatus.Name,
		"department_id":          item.DepartmentID,
		"department_name":        item.Department.Name,
		"job_id":                 item.JobID,
		"job_title":              item.Job.Title,
		"job_level_id":           item.Job.JobLevelID,
		"job_level_name":         item.Job.JobLevel.Name,
		"location_id":            item.LocationID,
		"location_name":          item.Location.Name,
		"work_mode":              item.WorkMode,
		"management_scope":       item.ManagementScope,
		"created_at":             item.CreatedAt,
		"updated_at":             item.UpdatedAt,
	}
	if item.Manager != nil {
		response["manager_employee_id"] = item.Manager.ID
		response["manager_name"] = shared.FullName(item.Manager.FirstName, item.Manager.LastName)
	}
	if item.UserID != nil {
		response["user_id"] = item.UserID
	}
	assignments := make([]map[string]any, 0, len(item.RoleAssignments))
	for _, assignment := range item.RoleAssignments {
		assignmentRow := map[string]any{
			"id":                       assignment.ID,
			"job_id":                   assignment.JobID,
			"job_title":                assignment.Job.Title,
			"job_level_name":           assignment.Job.JobLevel.Name,
			"department_id":            assignment.DepartmentID,
			"department_name":          assignment.Department.Name,
			"estimated_hours_per_week": assignment.EstimatedHoursPerWeek,
			"start_date":               assignment.StartDate.Format("2006-01-02"),
			"assignment_status":        shared.AssignmentStatus(assignment.StartDate, assignment.EndDate, time.Now()),
			"notes":                    assignment.Notes,
		}
		if assignment.EndDate != nil {
			assignmentRow["end_date"] = assignment.EndDate.Format("2006-01-02")
		}
		assignments = append(assignments, assignmentRow)
	}
	response["secondary_assignments"] = assignments
	return response
}

func normalizeWorkMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "remote":
		return "remote"
	case "hybrid":
		return "hybrid"
	case "client-based":
		return "client-based"
	default:
		return "onsite"
	}
}

func normalizeManagementScope(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case shared.ManagementScopeDivision:
		return shared.ManagementScopeDivision
	case shared.ManagementScopeSubdepartment:
		return shared.ManagementScopeSubdepartment
	case shared.ManagementScopeTeam:
		return shared.ManagementScopeTeam
	default:
		return shared.ManagementScopeIndividual
	}
}
