package profile

import (
	"errors"
	"net/http"
	"strings"

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

type updateRequest struct {
	Email       string  `json:"email"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	PhoneNumber *string `json:"phone_number"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/profile", func(r chi.Router) {
		r.Get("/", service.get)
		r.Put("/", service.update)
	})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	resp, err := s.loadProfile(claims)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		httpx.Error(w, http.StatusUnauthorized, "missing auth claims", nil)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "invalid auth user", nil)
		return
	}

	var req updateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if strings.TrimSpace(req.Email) == "" {
		httpx.Error(w, http.StatusBadRequest, "email is required", map[string]string{"email": "required"})
		return
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("email", strings.ToLower(strings.TrimSpace(req.Email))).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update user email", nil)
		return
	}

	var user models.User
	if err := s.db.Preload("Role").First(&user, "id = ?", userID).Error; err != nil {
		httpx.Error(w, http.StatusNotFound, "user not found", nil)
		return
	}

	employee, err := shared.FindLinkedEmployee(s.db, user)
	if err == nil && (req.FirstName != nil || req.LastName != nil || req.PhoneNumber != nil) {
		updates := map[string]any{}
		if req.FirstName != nil {
			updates["first_name"] = strings.TrimSpace(*req.FirstName)
		}
		if req.LastName != nil {
			updates["last_name"] = strings.TrimSpace(*req.LastName)
		}
		if req.PhoneNumber != nil {
			updates["phone_number"] = strings.TrimSpace(*req.PhoneNumber)
		}
		if len(updates) > 0 {
			if err := s.db.Model(&models.Employee{}).Where("id = ?", employee.ID).Updates(updates).Error; err != nil {
				httpx.Error(w, http.StatusBadRequest, "could not update employee profile", nil)
				return
			}
		}
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "profile", actorID, map[string]any{"email": req.Email})

	resp, err := s.loadProfile(claims)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s Service) loadProfile(claims *internalauth.AccessClaims) (map[string]any, error) {
	if claims == nil {
		return nil, errors.New("missing auth claims")
	}

	var user models.User
	if err := s.db.Preload("Role").First(&user, "id = ?", claims.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	resp := map[string]any{
		"id":              user.ID,
		"organization_id": user.OrganizationID,
		"email":           user.Email,
		"role":            user.Role.Code,
		"is_active":       user.IsActive,
	}

	employee, err := shared.FindLinkedEmployee(s.db, user)
	if err == nil {
		assignments := make([]map[string]any, 0, len(employee.RoleAssignments))
		for _, assignment := range employee.RoleAssignments {
			entry := map[string]any{
				"id":                       assignment.ID,
				"job_id":                   assignment.JobID,
				"job_title":                assignment.Job.Title,
				"job_level_name":           assignment.Job.JobLevel.Name,
				"department_id":            assignment.DepartmentID,
				"department_name":          assignment.Department.Name,
				"estimated_hours_per_week": assignment.EstimatedHoursPerWeek,
				"start_date":               assignment.StartDate.Format("2006-01-02"),
				"notes":                    assignment.Notes,
			}
			if assignment.EndDate != nil {
				entry["end_date"] = assignment.EndDate.Format("2006-01-02")
			}
			assignments = append(assignments, entry)
		}

		resp["employee"] = map[string]any{
			"id":                     employee.ID,
			"organization_id":        employee.OrganizationID,
			"employee_code":          employee.EmployeeCode,
			"first_name":             employee.FirstName,
			"last_name":              employee.LastName,
			"full_name":              shared.FullName(employee.FirstName, employee.LastName),
			"phone_number":           employee.PhoneNumber,
			"department_id":          employee.DepartmentID,
			"department_name":        employee.Department.Name,
			"job_id":                 employee.JobID,
			"job_title":              employee.Job.Title,
			"job_level_id":           employee.Job.JobLevelID,
			"job_level_name":         employee.Job.JobLevel.Name,
			"location_id":            employee.LocationID,
			"location_name":          employee.Location.Name,
			"work_mode":              employee.WorkMode,
			"management_scope":       employee.ManagementScope,
			"employee_type_id":       employee.EmployeeTypeID,
			"employee_type_name":     employee.EmployeeType.Name,
			"employment_status_id":   employee.EmploymentStatusID,
			"employment_status_name": employee.EmploymentStatus.Name,
			"manager_employee_id":    employee.ManagerEmployeeID,
			"manager_name":           managerName(employee.Manager),
			"secondary_assignments":  assignments,
		}
	}

	return resp, nil
}

func managerName(manager *models.Employee) string {
	if manager == nil {
		return ""
	}
	return shared.FullName(manager.FirstName, manager.LastName)
}
