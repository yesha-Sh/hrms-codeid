package attendances

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
	EmployeeID     string  `json:"employee_id"`
	AttendanceDate string  `json:"attendance_date"`
	CheckInAt      *string `json:"check_in_at"`
	CheckOutAt     *string `json:"check_out_at"`
	Status         string  `json:"status"`
	Notes          string  `json:"notes"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/attendances", func(r chi.Router) {
		r.Get("/", service.list)
		r.Post("/", service.create)
		r.Get("/{id}", service.get)
		r.Put("/{id}", service.update)
		r.Delete("/{id}", service.delete)
	})
}

func (s Service) list(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	pagination := shared.ParsePagination(r, "attendance_date")
	query := s.baseQuery()

	if search := r.URL.Query().Get("search"); search != "" {
		term := "%" + search + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR employees.employee_code ILIKE ? OR attendances.status ILIKE ?", term, term, term, term)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("attendances.status = ?", status)
	}
	if date := r.URL.Query().Get("date"); date != "" {
		query = query.Where("attendances.attendance_date = ?", date)
	}
	if locationID := r.URL.Query().Get("location_id"); locationID != "" {
		query = query.Where("employees.location_id = ?", locationID)
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		query = query.Where("attendances.employee_id = ?", employeeID)
	}
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("employees.department_id = ?", departmentID)
	}

	scoped, err := s.applyScope(query, claims)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var total int64
	if err := scoped.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count attendances", nil)
		return
	}

	var attendances []models.Attendance
	if err := scoped.Order("attendances." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&attendances).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list attendances", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"items": toListResponse(attendances),
		"meta":  map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total},
	})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid attendance id", nil)
		return
	}

	attendance, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "attendance not found", nil)
		return
	}
	if err := s.ensureAccess(attendance, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	httpx.JSON(w, http.StatusOK, toResponse(attendance))
}

func (s Service) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	attendance, err := s.toModel(req, claims)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var existingCount int64
	if err := s.db.Model(&models.Attendance{}).Where("organization_id = ? AND employee_id = ? AND attendance_date = ?", attendance.OrganizationID, attendance.EmployeeID, attendance.AttendanceDate).Count(&existingCount).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not validate attendance uniqueness", nil)
		return
	}
	if existingCount > 0 {
		httpx.Error(w, http.StatusBadRequest, "attendance for this employee and date already exists", nil)
		return
	}

	if err := s.db.Create(&attendance).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create attendance", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "attendance", &attendance.ID, map[string]any{"status": attendance.Status})

	fresh, _ := s.findByID(attendance.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid attendance id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "attendance not found", nil)
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

	if err := s.db.Model(&current).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update attendance", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "attendance", &current.ID, map[string]any{"status": updated.Status})

	fresh, _ := s.findByID(current.ID)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid attendance id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "attendance not found", nil)
		return
	}
	if err := s.ensureAccess(current, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	if err := s.db.Delete(&models.Attendance{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete attendance", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "attendance", &id, nil)
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "attendance deleted"})
}

func (s Service) baseQuery() *gorm.DB {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return s.db.Model(&models.Attendance{}).Where("1 = 0")
	}

	return s.db.Model(&models.Attendance{}).
		Joins("JOIN employees ON employees.id = attendances.employee_id").
		Where("attendances.organization_id = ?", organizationID).
		Preload("Employee.Department").
		Preload("Employee.Location").
		Preload("Employee.Manager")
}

func (s Service) applyScope(query *gorm.DB, claims *internalauth.AccessClaims) (*gorm.DB, error) {
	switch claims.Role {
	case "admin":
		return query, nil
	case "manager":
		managerID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || managerID == nil {
			return nil, errors.New("manager employee context is required")
		}
		context, err := shared.LoadManagerContext(s.db, *managerID)
		if err != nil {
			return nil, err
		}
		if context.ManagementScope == shared.ManagementScopeTeam {
			ids := append([]uuid.UUID{*managerID}, context.ManagedTeamMemberIDs...)
			return query.Where("attendances.employee_id IN ?", ids), nil
		}
		return query.Where("attendances.employee_id = ?", *managerID), nil
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return nil, errors.New("employee context is required")
		}
		return query.Where("attendances.employee_id = ?", *employeeID), nil
	default:
		return nil, errors.New("unsupported role")
	}
}

func (s Service) ensureAccess(item models.Attendance, claims *internalauth.AccessClaims) error {
	switch claims.Role {
	case "admin":
		return nil
	case "manager":
		managerID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || managerID == nil {
			return errors.New("manager employee context is required")
		}
		if item.EmployeeID == *managerID {
			return nil
		}
		return errors.New("managers can only change their own attendance records")
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return errors.New("employee context is required")
		}
		if item.EmployeeID != *employeeID {
			return errors.New("attendance is outside user scope")
		}
		return nil
	default:
		return errors.New("unsupported role")
	}
}

func (s Service) findByID(id uuid.UUID) (models.Attendance, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Attendance{}, err
	}
	var attendance models.Attendance
	err = s.baseQuery().Preload("Employee.Manager").First(&attendance, "attendances.id = ? AND attendances.organization_id = ?", id, organizationID).Error
	return attendance, err
}

func (s Service) toModel(req upsertRequest, claims *internalauth.AccessClaims) (models.Attendance, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Attendance{}, err
	}
	dateValue, err := time.Parse("2006-01-02", req.AttendanceDate)
	if err != nil {
		return models.Attendance{}, err
	}

	employeeID, err := s.resolveEmployeeID(req.EmployeeID, claims)
	if err != nil {
		return models.Attendance{}, err
	}

	var employee models.Employee
	if err := s.db.Preload("EmployeeType").First(&employee, "id = ? AND organization_id = ?", employeeID, organizationID).Error; err != nil {
		return models.Attendance{}, err
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "on time"
	}
	if shared.RequiresRemoteAttendance(employee.WorkMode, employee.EmployeeType.Code) {
		status = "remote"
	}

	attendance := models.Attendance{
		OrganizationID: organizationID,
		EmployeeID:     employeeID,
		AttendanceDate: dateValue,
		Status:         status,
		Notes:          req.Notes,
	}

	if req.CheckInAt != nil && *req.CheckInAt != "" {
		checkIn, err := time.Parse(time.RFC3339, *req.CheckInAt)
		if err != nil {
			return models.Attendance{}, err
		}
		attendance.CheckInAt = &checkIn
	}
	if req.CheckOutAt != nil && *req.CheckOutAt != "" {
		checkOut, err := time.Parse(time.RFC3339, *req.CheckOutAt)
		if err != nil {
			return models.Attendance{}, err
		}
		attendance.CheckOutAt = &checkOut
	}

	return attendance, nil
}

func (s Service) resolveEmployeeID(raw string, claims *internalauth.AccessClaims) (uuid.UUID, error) {
	var employeeID uuid.UUID
	switch claims.Role {
	case "employee":
		value, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || value == nil {
			return uuid.Nil, errors.New("employee context is required")
		}
		employeeID = *value
	case "manager":
		value, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || value == nil {
			return uuid.Nil, errors.New("manager employee context is required")
		}
		if raw != "" {
			parsed, err := uuid.Parse(raw)
			if err != nil {
				return uuid.Nil, err
			}
			if parsed != *value {
				return uuid.Nil, errors.New("managers can only create attendance for themselves")
			}
		}
		employeeID = *value
	default:
		var err error
		employeeID, err = uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, err
		}
	}

	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return uuid.Nil, err
	}

	var employee models.Employee
	if err := s.db.Select("id", "manager_employee_id", "department_id").First(&employee, "id = ? AND organization_id = ?", employeeID, organizationID).Error; err != nil {
		return uuid.Nil, err
	}
	return employeeID, nil
}

func toListResponse(items []models.Attendance) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.Attendance) map[string]any {
	return map[string]any{
		"id":              item.ID,
		"organization_id": item.OrganizationID,
		"employee_id":     item.EmployeeID,
		"employee_name":   shared.FullName(item.Employee.FirstName, item.Employee.LastName),
		"employee_code":   item.Employee.EmployeeCode,
		"department_id":   item.Employee.DepartmentID,
		"department_name": item.Employee.Department.Name,
		"location_id":     item.Employee.LocationID,
		"location_name":   item.Employee.Location.Name,
		"work_mode":       item.Employee.WorkMode,
		"attendance_date": item.AttendanceDate.Format("2006-01-02"),
		"check_in_at":     item.CheckInAt,
		"check_out_at":    item.CheckOutAt,
		"status":          item.Status,
		"notes":           item.Notes,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	}
}
