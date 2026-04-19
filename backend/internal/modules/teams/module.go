package teams

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

type addMemberRequest struct {
	EmployeeID string `json:"employee_id"`
	RoleName   string `json:"role_name"`
	Notes      string `json:"notes"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/teams", func(r chi.Router) {
		r.Get("/", service.list)
		r.Get("/{id}/members", service.members)
		r.Get("/{id}/available-employees", service.availableEmployees)
		r.Post("/{id}/members", service.addMember)
		r.Delete("/{id}/members/{membershipId}", service.removeMember)
	})
}

func (s Service) list(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	query := s.baseQuery()
	scoped, err := s.applyScope(query, claims)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var teams []models.Team
	if err := scoped.Order("teams.name asc").Find(&teams).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list teams", nil)
		return
	}

	items := make([]map[string]any, 0, len(teams))
	for _, team := range teams {
		var activeMembers int64
		_ = s.db.Model(&models.TeamMembership{}).
			Where("team_id = ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", team.ID).
			Count(&activeMembers).Error

		items = append(items, toTeamResponse(team, activeMembers))
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s Service) members(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	team, err := s.findTeam(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "team not found", nil)
		return
	}
	if err := s.ensureTeamAccess(team, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	query := s.db.Model(&models.TeamMembership{}).
		Where("team_id = ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", team.ID).
		Preload("Employee.Department").
		Preload("Employee.Job.JobLevel").
		Preload("Employee.Location").
		Order("employees.first_name asc")

	var memberships []models.TeamMembership
	if err := query.Joins("JOIN employees ON employees.id = team_memberships.employee_id").Find(&memberships).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load team members", nil)
		return
	}

	items := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		items = append(items, toMembershipResponse(membership))
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"team": toTeamResponse(team, int64(len(items))), "items": items})
}

func (s Service) availableEmployees(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	team, err := s.findTeam(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "team not found", nil)
		return
	}
	if err := s.ensureTeamAccess(team, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	query := shared.PreloadEmployeeDetails(s.db.Model(&models.Employee{})).
		Where("employees.organization_id = ?", organizationID).
		Joins("JOIN departments scoped_departments ON scoped_departments.id = employees.department_id").
		Joins("JOIN employment_statuses ON employment_statuses.id = employees.employment_status_id").
		Joins("LEFT JOIN users team_users ON team_users.id = employees.user_id").
		Joins("LEFT JOIN roles team_roles ON team_roles.id = team_users.role_id").
		Where("employment_statuses.code IN ?", []string{"active", "probation"}).
		Where("scoped_departments.level > 0").
		Where("employees.management_scope = ?", shared.ManagementScopeIndividual).
		Where("(team_roles.code IS NULL OR team_roles.code = ?)", "employee")

	if !team.IsCrossDepartment && team.DepartmentID != nil {
		query = query.Where("employees.department_id = ?", *team.DepartmentID)
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		term := "%" + search + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR employees.employee_code ILIKE ? OR employees.email ILIKE ?", term, term, term, term)
	}

	// Keep team membership simple in v1: an employee can only have one active team.
	subQuery := s.db.Model(&models.TeamMembership{}).
		Select("employee_id").
		Where("organization_id = ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", organizationID)
	query = query.Where("employees.id NOT IN (?)", subQuery)
	query = query.Where("employees.id <> ?", team.ManagerEmployeeID)

	var employees []models.Employee
	if err := query.Order("employees.first_name asc, employees.last_name asc").Find(&employees).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load available employees", nil)
		return
	}

	items := make([]map[string]any, 0, len(employees))
	for _, employee := range employees {
		items = append(items, map[string]any{
			"id":                employee.ID,
			"employee_code":     employee.EmployeeCode,
			"full_name":         shared.FullName(employee.FirstName, employee.LastName),
			"department_name":   employee.Department.Name,
			"job_title":         employee.Job.Title,
			"job_level_name":    employee.Job.JobLevel.Name,
			"employment_status": employee.EmploymentStatus.Name,
			"location_name":     employee.Location.Name,
			"work_mode":         employee.WorkMode,
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"team": toTeamResponse(team, 0), "items": items})
}

func (s Service) addMember(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	team, err := s.findTeam(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "team not found", nil)
		return
	}
	if err := s.ensureTeamAccess(team, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var req addMemberRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid employee id", nil)
		return
	}

	employee, err := s.findEmployee(employeeID)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "employee not found", nil)
		return
	}
	if !team.IsCrossDepartment && team.DepartmentID != nil && employee.DepartmentID != *team.DepartmentID {
		httpx.Error(w, http.StatusBadRequest, "employee must belong to the team's department", nil)
		return
	}

	var activeMembership models.TeamMembership
	existing := s.db.Where("organization_id = ? AND employee_id = ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", employee.OrganizationID, employee.ID).First(&activeMembership)
	if existing.Error == nil {
		httpx.Error(w, http.StatusBadRequest, "employee already belongs to an active team", nil)
		return
	}
	if existing.Error != nil && !errors.Is(existing.Error, gorm.ErrRecordNotFound) {
		httpx.Error(w, http.StatusInternalServerError, "could not validate team membership", nil)
		return
	}

	roleName := strings.TrimSpace(req.RoleName)
	if roleName == "" {
		roleName = "Member"
	}

	startDate := time.Now().UTC().Truncate(24 * time.Hour)
	var membership models.TeamMembership
	lookup := s.db.Where("team_id = ? AND employee_id = ?", team.ID, employee.ID).First(&membership)
	switch {
	case lookup.Error == nil:
		if err := s.db.Model(&models.TeamMembership{}).Where("id = ?", membership.ID).Updates(map[string]any{
			"organization_id": team.OrganizationID,
			"role_name":       roleName,
			"start_date":      startDate,
			"end_date":        nil,
			"notes":           strings.TrimSpace(req.Notes),
		}).Error; err != nil {
			httpx.Error(w, http.StatusBadRequest, "could not add team member", nil)
			return
		}
	case errors.Is(lookup.Error, gorm.ErrRecordNotFound):
		membership = models.TeamMembership{
			OrganizationID: team.OrganizationID,
			TeamID:         team.ID,
			EmployeeID:     employee.ID,
			RoleName:       roleName,
			StartDate:      startDate,
			Notes:          strings.TrimSpace(req.Notes),
		}
		if err := s.db.Create(&membership).Error; err != nil {
			httpx.Error(w, http.StatusBadRequest, "could not add team member", nil)
			return
		}
	default:
		httpx.Error(w, http.StatusInternalServerError, "could not validate team membership history", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "team_membership", &membership.ID, map[string]any{"team_id": team.ID, "employee_id": employee.ID})

	fresh, err := s.findMembership(membership.ID)
	if err != nil {
		httpx.Error(w, http.StatusCreated, "team member added", nil)
		return
	}
	httpx.JSON(w, http.StatusCreated, toMembershipResponse(fresh))
}

func (s Service) removeMember(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	team, err := s.findTeam(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "team not found", nil)
		return
	}
	if err := s.ensureTeamAccess(team, claims); err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	membershipID, err := uuid.Parse(chi.URLParam(r, "membershipId"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid membership id", nil)
		return
	}

	membership, err := s.findMembership(membershipID)
	if err != nil || membership.TeamID != team.ID {
		httpx.Error(w, http.StatusNotFound, "team membership not found", nil)
		return
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.TeamMembership{}).Where("id = ?", membership.ID).Updates(map[string]any{
		"end_date":    now,
		"updated_at":  time.Now().UTC(),
	}).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not remove team member", nil)
		return
	}

	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "team_membership", &membership.ID, map[string]any{"team_id": team.ID, "employee_id": membership.EmployeeID, "end_date": now.Format("2006-01-02")})
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "team member removed"})
}

func (s Service) baseQuery() *gorm.DB {
	organizationID, _ := shared.MustOrganizationID(s.db)
	return s.db.Model(&models.Team{}).
		Where("teams.organization_id = ?", organizationID).
		Preload("Department").
		Preload("ManagerEmployee.Department")
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
		return query.Where("teams.manager_employee_id = ?", *managerID), nil
	default:
		return nil, errors.New("team access is limited to admins and managers")
	}
}

func (s Service) ensureTeamAccess(team models.Team, claims *internalauth.AccessClaims) error {
	if claims == nil {
		return errors.New("missing auth claims")
	}
	if claims.Role == "admin" {
		return nil
	}
	if claims.Role != "manager" {
		return errors.New("team access is limited to admins and managers")
	}
	managerID, err := shared.CurrentEmployeeUUID(claims)
	if err != nil || managerID == nil {
		return errors.New("manager employee context is required")
	}
	if team.ManagerEmployeeID != *managerID {
		return errors.New("team is outside manager scope")
	}
	return nil
}

func (s Service) findTeam(rawID string) (models.Team, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return models.Team{}, err
	}
	var team models.Team
	err = s.baseQuery().First(&team, "teams.id = ?", id).Error
	return team, err
}

func (s Service) findEmployee(id uuid.UUID) (models.Employee, error) {
	var employee models.Employee
	err := shared.PreloadEmployeeDetails(s.db.Model(&models.Employee{})).First(&employee, "employees.id = ?", id).Error
	return employee, err
}

func (s Service) findMembership(id uuid.UUID) (models.TeamMembership, error) {
	var membership models.TeamMembership
	err := s.db.Model(&models.TeamMembership{}).
		Preload("Employee.Department").
		Preload("Employee.Job.JobLevel").
		Preload("Employee.Location").
		Preload("Team.Department").
		First(&membership, "team_memberships.id = ?", id).Error
	return membership, err
}

func toTeamResponse(team models.Team, memberCount int64) map[string]any {
	response := map[string]any{
		"id":                  team.ID,
		"name":                team.Name,
		"manager_employee_id": team.ManagerEmployeeID,
		"manager_name":        shared.FullName(team.ManagerEmployee.FirstName, team.ManagerEmployee.LastName),
		"is_cross_department": team.IsCrossDepartment,
		"focus_area":          team.FocusArea,
		"member_count":        memberCount,
	}
	if team.Department != nil {
		response["department_id"] = team.DepartmentID
		response["department_name"] = team.Department.Name
	}
	return response
}

func toMembershipResponse(membership models.TeamMembership) map[string]any {
	response := map[string]any{
		"id":                       membership.ID,
		"team_id":                  membership.TeamID,
		"employee_id":              membership.EmployeeID,
		"employee_code":            membership.Employee.EmployeeCode,
		"employee_name":            shared.FullName(membership.Employee.FirstName, membership.Employee.LastName),
		"department_name":          membership.Employee.Department.Name,
		"job_title":                membership.Employee.Job.Title,
		"job_level_name":           membership.Employee.Job.JobLevel.Name,
		"location_name":            membership.Employee.Location.Name,
		"work_mode":                membership.Employee.WorkMode,
		"role_name":                membership.RoleName,
		"start_date":               membership.StartDate.Format("2006-01-02"),
		"notes":                    membership.Notes,
	}
	if membership.EndDate != nil {
		response["end_date"] = membership.EndDate.Format("2006-01-02")
	}
	return response
}
