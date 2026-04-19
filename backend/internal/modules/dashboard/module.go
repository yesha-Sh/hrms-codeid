package dashboard

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"hrms/backend/internal/httpx"
	"hrms/backend/internal/middleware"
	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/shared"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return Service{db: db}
}

func Mount(router chi.Router, service Service) {
	router.Route("/dashboard", func(r chi.Router) {
		r.Get("/admin", service.admin)
		r.Get("/manager", service.manager)
		r.Get("/employee", service.employee)
	})
}

func (s Service) admin(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, "admin access required", nil)
		return
	}

	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	var employees int64
	var departments int64
	var pendingLeaves int64
	var presentToday int64
	var activeAssignments int64
	var locations int64

	s.db.Model(&models.Employee{}).Where("organization_id = ?", organizationID).Count(&employees)
	s.db.Model(&models.Department{}).Where("organization_id = ?", organizationID).Count(&departments)
	s.db.Model(&models.LeaveRequest{}).Where("organization_id = ? AND status = ?", organizationID, "pending").Count(&pendingLeaves)
	s.db.Model(&models.Attendance{}).Where("organization_id = ? AND attendance_date = CURRENT_DATE", organizationID).Count(&presentToday)
	s.db.Model(&models.EmployeeRoleAssignment{}).Where("organization_id = ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", organizationID).Count(&activeAssignments)
	s.db.Model(&models.Location{}).Where("organization_id = ?", organizationID).Count(&locations)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"organization_name":            "PT. CODEID",
		"total_employees":              employees,
		"total_departments":            departments,
		"total_locations":              locations,
		"pending_leave_requests":       pendingLeaves,
		"today_attendance":             presentToday,
		"active_secondary_assignments": activeAssignments,
	})
}

func (s Service) manager(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "manager" {
		httpx.Error(w, http.StatusForbidden, "manager access required", nil)
		return
	}

	managerID, err := shared.CurrentEmployeeUUID(claims)
	if err != nil || managerID == nil {
		httpx.Error(w, http.StatusForbidden, "manager employee context is required", nil)
		return
	}

	context, err := shared.LoadManagerContext(s.db, *managerID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not resolve manager scope", nil)
		return
	}

	var teamMemberIDs []uuid.UUID
	switch context.ManagementScope {
	case shared.ManagementScopeDivision, shared.ManagementScopeSubdepartment:
		if len(context.ManagedDepartmentIDs) > 0 {
			if err := s.db.Model(&models.Employee{}).
				Where("department_id IN ?", context.ManagedDepartmentIDs).
				Pluck("id", &teamMemberIDs).Error; err != nil {
				httpx.Error(w, http.StatusInternalServerError, "could not resolve department members", nil)
				return
			}
		}
	case shared.ManagementScopeTeam:
		teamMemberIDs = append(teamMemberIDs, context.ManagedTeamMemberIDs...)
	}

	var presentToday int64
	var pendingApprovals int64
	var activeAssignments int64

	if len(teamMemberIDs) > 0 {
		s.db.Model(&models.Attendance{}).Where("employee_id IN ? AND attendance_date = CURRENT_DATE", teamMemberIDs).Count(&presentToday)
		s.db.Model(&models.LeaveRequest{}).Where("employee_id IN ? AND status = ?", teamMemberIDs, "pending").Count(&pendingApprovals)
		s.db.Model(&models.EmployeeRoleAssignment{}).Where("employee_id IN ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", teamMemberIDs).Count(&activeAssignments)
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"management_scope":             context.ManagementScope,
		"team_members":                 len(teamMemberIDs),
		"today_present":                presentToday,
		"pending_approvals":            pendingApprovals,
		"active_secondary_assignments": activeAssignments,
		"managed_teams":                len(context.ManagedTeamIDs),
		"managed_departments":          len(context.ManagedDepartmentRoots),
	})
}

func (s Service) employee(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != "employee" {
		httpx.Error(w, http.StatusForbidden, "employee access required", nil)
		return
	}

	employeeID, err := shared.CurrentEmployeeUUID(claims)
	if err != nil || employeeID == nil {
		httpx.Error(w, http.StatusForbidden, "employee context is required", nil)
		return
	}

	var attendanceCount int64
	var pendingLeaves int64
	var activeAssignments int64
	var employee models.Employee

	s.db.Model(&models.Attendance{}).Where("employee_id = ?", *employeeID).Count(&attendanceCount)
	s.db.Model(&models.LeaveRequest{}).Where("employee_id = ? AND status = ?", *employeeID, "pending").Count(&pendingLeaves)
	s.db.Model(&models.EmployeeRoleAssignment{}).Where("employee_id = ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", *employeeID).Count(&activeAssignments)
	s.db.Preload("Job.JobLevel").First(&employee, "id = ?", *employeeID)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"attendance_entries":           attendanceCount,
		"pending_leaves":               pendingLeaves,
		"active_secondary_assignments": activeAssignments,
		"job_level_name":               employee.Job.JobLevel.Name,
		"work_mode":                    employee.WorkMode,
	})
}
