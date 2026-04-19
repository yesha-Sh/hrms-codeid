package leave

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
	EmployeeID         *string `json:"employee_id"`
	ApproverEmployeeID *string `json:"approver_employee_id"`
	LeaveTypeID        string  `json:"leave_type_id"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
	Reason             string  `json:"reason"`
	Status             string  `json:"status"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/leave-requests", func(r chi.Router) {
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
		query = query.Joins("LEFT JOIN leave_types ON leave_types.id = leave_requests.leave_type_id").
			Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR employees.employee_code ILIKE ? OR leave_types.name ILIKE ? OR leave_requests.status ILIKE ?", term, term, term, term, term)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("leave_requests.status = ?", status)
	}
	if leaveTypeID := r.URL.Query().Get("leave_type_id"); leaveTypeID != "" {
		query = query.Where("leave_requests.leave_type_id = ?", leaveTypeID)
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		query = query.Where("leave_requests.employee_id = ?", employeeID)
	}
	if approverID := r.URL.Query().Get("approver_id"); approverID != "" {
		query = query.Where("leave_requests.approver_employee_id = ?", approverID)
	}
	if mine := r.URL.Query().Get("mine"); mine == "true" {
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			httpx.Error(w, http.StatusForbidden, "employee context is required", nil)
			return
		}
		query = query.Where("leave_requests.employee_id = ?", *employeeID)
	}

	scoped, err := s.applyScope(query, claims)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var total int64
	if err := scoped.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count leave requests", nil)
		return
	}

	var items []models.LeaveRequest
	if err := scoped.Order("leave_requests." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&items).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list leave requests", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": toListResponse(items), "meta": map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total}})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid leave request id", nil)
		return
	}

	item, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "leave request not found", nil)
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

	if claims != nil && (claims.Role == "employee" || claims.Role == "manager") {
		item.Status = "pending"
		item.ApprovedAt = nil
		if claims.Role == "employee" {
			item.ApproverEmployeeID = nil
		} else {
			approverID, err := s.resolveManagerSelfLeaveApprover(item.EmployeeID)
			if err != nil {
				httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
				return
			}
			item.ApproverEmployeeID = approverID
		}
	}

	if err := s.db.Create(&item).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create leave request", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "leave_request", &item.ID, map[string]any{"status": item.Status})

	fresh, _ := s.findByID(item.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid leave request id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "leave request not found", nil)
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

	if claims != nil {
		switch claims.Role {
		case "employee":
			if err := validateSelfServiceLeaveStatus(current, req.Status); err != nil {
				httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
				return
			}
		case "manager":
			managerID, err := shared.CurrentEmployeeUUID(claims)
			if err != nil || managerID == nil {
				httpx.Error(w, http.StatusForbidden, "manager employee context is required", nil)
				return
			}
			if current.EmployeeID == *managerID {
				if err := validateSelfServiceLeaveStatus(current, req.Status); err != nil {
					httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
					return
				}
			} else if !isPendingReviewStatus(current.Status) {
				httpx.Error(w, http.StatusBadRequest, "only pending or in-review leave requests can be decided", nil)
				return
			} else if !isApprovalStatus(req.Status) {
				httpx.Error(w, http.StatusBadRequest, "managers can only approve or reject team leave requests", nil)
				return
			}
		}
	}

	updated, err := s.toModel(req, claims)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	updated.ID = current.ID
	updated.OrganizationID = current.OrganizationID

	if claims != nil && claims.Role == "employee" {
		updated.EmployeeID = current.EmployeeID
		updated.ApproverEmployeeID = current.ApproverEmployeeID
		updated.ApprovedAt = current.ApprovedAt
	}

	if claims != nil && claims.Role == "manager" {
		managerID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || managerID == nil {
			httpx.Error(w, http.StatusForbidden, "manager employee context is required", nil)
			return
		}

		if current.EmployeeID == *managerID {
			updated.EmployeeID = current.EmployeeID
			approverID, err := s.resolveManagerSelfLeaveApprover(current.EmployeeID)
			if err != nil {
				httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
				return
			}
			updated.ApproverEmployeeID = approverID
			updated.ApprovedAt = current.ApprovedAt
		} else {
			updated.EmployeeID = current.EmployeeID
			if isApprovalStatus(updated.Status) {
				now := time.Now().UTC()
				updated.ApprovedAt = &now
				updated.ApproverEmployeeID = managerID
			} else {
				updated.ApproverEmployeeID = current.ApproverEmployeeID
				updated.ApprovedAt = current.ApprovedAt
			}
		}
	}

	if claims != nil && claims.Role == "admin" && isApprovalStatus(updated.Status) {
		now := time.Now().UTC()
		updated.ApprovedAt = &now
		if approverID, err := shared.CurrentEmployeeUUID(claims); err == nil && approverID != nil {
			updated.ApproverEmployeeID = approverID
		}
	}

	if err := s.db.Model(&current).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update leave request", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "leave_request", &current.ID, map[string]any{"status": updated.Status})

	fresh, _ := s.findByID(current.ID)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid leave request id", nil)
		return
	}

	current, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "leave request not found", nil)
		return
	}
	if err := s.ensureAccess(current, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	if claims != nil {
		switch claims.Role {
		case "employee":
			if isFinalLeaveStatus(current.Status) {
				httpx.Error(w, http.StatusBadRequest, "finalized leave requests cannot be deleted", nil)
				return
			}
		case "manager":
			managerID, err := shared.CurrentEmployeeUUID(claims)
			if err != nil || managerID == nil {
				httpx.Error(w, http.StatusForbidden, "manager employee context is required", nil)
				return
			}
			if current.EmployeeID != *managerID {
				httpx.Error(w, http.StatusForbidden, "managers cannot delete team leave requests", nil)
				return
			}
			if isFinalLeaveStatus(current.Status) {
				httpx.Error(w, http.StatusBadRequest, "finalized leave requests cannot be deleted", nil)
				return
			}
		}
	}

	if err := s.db.Delete(&models.LeaveRequest{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete leave request", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "leave_request", &id, nil)
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "leave request deleted"})
}

func (s Service) baseQuery() *gorm.DB {
	organizationID, _ := shared.MustOrganizationID(s.db)
	return s.db.Model(&models.LeaveRequest{}).
		Joins("JOIN employees ON employees.id = leave_requests.employee_id").
		Preload("Employee.Department").
		Preload("Approver").
		Preload("LeaveType").
		Where("leave_requests.organization_id = ?", organizationID)
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
		switch context.ManagementScope {
		case shared.ManagementScopeDivision, shared.ManagementScopeSubdepartment:
			if len(context.ManagedDepartmentIDs) == 0 {
				return query.Where("leave_requests.employee_id = ?", *managerID), nil
			}
			return query.Where("leave_requests.employee_id = ? OR employees.department_id IN ?", *managerID, context.ManagedDepartmentIDs), nil
		default:
			return query.Where("leave_requests.employee_id = ?", *managerID), nil
		}
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return nil, errors.New("employee context is required")
		}
		return query.Where("leave_requests.employee_id = ?", *employeeID), nil
	default:
		return nil, errors.New("unsupported role")
	}
}

func (s Service) ensureAccess(item models.LeaveRequest, claims *internalauth.AccessClaims) error {
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
		context, err := shared.LoadManagerContext(s.db, *managerID)
		if err != nil {
			return err
		}
		if !shared.EmployeeWithinManagerContext(item.Employee, context) {
			return errors.New("leave request is outside manager scope")
		}
		return nil
	case "employee":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return errors.New("employee context is required")
		}
		if item.EmployeeID != *employeeID {
			return errors.New("leave request is outside user scope")
		}
		return nil
	default:
		return errors.New("unsupported role")
	}
}

func (s Service) findByID(id uuid.UUID) (models.LeaveRequest, error) {
	var item models.LeaveRequest
	err := s.baseQuery().Preload("Employee.Manager").First(&item, "leave_requests.id = ?", id).Error
	return item, err
}

func (s Service) toModel(req upsertRequest, claims *internalauth.AccessClaims) (models.LeaveRequest, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return models.LeaveRequest{}, err
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return models.LeaveRequest{}, err
	}
	leaveTypeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return models.LeaveRequest{}, err
	}
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.LeaveRequest{}, err
	}
	employeeID, err := s.resolveEmployeeID(req.EmployeeID, claims)
	if err != nil {
		return models.LeaveRequest{}, err
	}

	item := models.LeaveRequest{OrganizationID: organizationID, EmployeeID: employeeID, LeaveTypeID: leaveTypeID, StartDate: startDate, EndDate: endDate, Reason: req.Reason, Status: req.Status}
	if item.Status == "" {
		item.Status = "pending"
	}

	if req.ApproverEmployeeID != nil && *req.ApproverEmployeeID != "" && claims != nil && claims.Role == "admin" {
		value, err := uuid.Parse(*req.ApproverEmployeeID)
		if err != nil {
			return models.LeaveRequest{}, err
		}
		item.ApproverEmployeeID = &value
	}

	return item, nil
}

func (s Service) resolveEmployeeID(raw *string, claims *internalauth.AccessClaims) (uuid.UUID, error) {
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
		employeeID = *value
	default:
		if raw == nil || *raw == "" {
			return uuid.Nil, errors.New("employee_id is required")
		}
		value, err := uuid.Parse(*raw)
		if err != nil {
			return uuid.Nil, err
		}
		employeeID = value
	}

	var employee models.Employee
	if err := s.db.Select("id", "department_id", "manager_employee_id").First(&employee, "id = ?", employeeID).Error; err != nil {
		return uuid.Nil, err
	}

	if claims.Role == "manager" {
		managerID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || managerID == nil {
			return uuid.Nil, errors.New("manager employee context is required")
		}
		if employee.ID == *managerID {
			return employeeID, nil
		}
		context, err := shared.LoadManagerContext(s.db, *managerID)
		if err != nil {
			return uuid.Nil, err
		}
		if !shared.EmployeeWithinManagerContext(employee, context) {
			return uuid.Nil, errors.New("leave request is outside manager scope")
		}
	}

	return employeeID, nil
}

func (s Service) resolveManagerSelfLeaveApprover(employeeID uuid.UUID) (*uuid.UUID, error) {
	approverID, err := shared.FindNearestParentManagerID(s.db, employeeID)
	if err != nil {
		return nil, err
	}
	if approverID != nil {
		return approverID, nil
	}

	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return nil, err
	}
	return shared.FindAdminApproverEmployeeID(s.db, organizationID)
}

func (s Service) managerScope(claims *internalauth.AccessClaims) (uuid.UUID, []uuid.UUID, error) {
	managerID, err := shared.CurrentEmployeeUUID(claims)
	if err != nil || managerID == nil {
		return uuid.Nil, nil, errors.New("manager employee context is required")
	}
	departmentIDs, err := shared.AccessibleDepartmentIDs(s.db, *managerID)
	if err != nil {
		return uuid.Nil, nil, err
	}
	return *managerID, departmentIDs, nil
}

func isApprovalStatus(status string) bool {
	return status == "approved" || status == "rejected"
}

func isPendingReviewStatus(status string) bool {
	return status == "pending" || status == "review"
}

func isFinalLeaveStatus(status string) bool {
	return status == "approved" || status == "rejected"
}

func validateSelfServiceLeaveStatus(current models.LeaveRequest, nextStatus string) error {
	if current.Status == "cancelled" && nextStatus != "cancelled" {
		return errors.New("cancelled leave requests cannot be reopened")
	}
	switch nextStatus {
	case "", "pending":
		if current.Status == "approved" || current.Status == "rejected" {
			return errors.New("finalized leave requests cannot be changed")
		}
		return nil
	case "cancelled":
		if current.Status == "approved" || current.Status == "rejected" {
			return errors.New("finalized leave requests cannot be cancelled")
		}
		return nil
	default:
		return errors.New("self-service leave requests can only stay pending or be cancelled")
	}
}

func toListResponse(items []models.LeaveRequest) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.LeaveRequest) map[string]any {
	response := map[string]any{"id": item.ID, "organization_id": item.OrganizationID, "employee_id": item.EmployeeID, "employee_name": shared.FullName(item.Employee.FirstName, item.Employee.LastName), "employee_code": item.Employee.EmployeeCode, "department_name": item.Employee.Department.Name, "leave_type_id": item.LeaveTypeID, "leave_type_name": item.LeaveType.Name, "start_date": item.StartDate.Format("2006-01-02"), "end_date": item.EndDate.Format("2006-01-02"), "reason": item.Reason, "status": item.Status, "approved_at": item.ApprovedAt, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
	if item.Approver != nil {
		response["approver_employee_id"] = item.Approver.ID
		response["approver_name"] = shared.FullName(item.Approver.FirstName, item.Approver.LastName)
	}
	return response
}
