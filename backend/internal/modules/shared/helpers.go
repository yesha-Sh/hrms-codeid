package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	internalauth "hrms/backend/internal/auth"
	"hrms/backend/internal/models"
)

const DefaultOrganizationCode = "pt-codeid"

const (
	ManagementScopeIndividual    = "individual_contributor"
	ManagementScopeDivision      = "division_manager"
	ManagementScopeSubdepartment = "subdepartment_manager"
	ManagementScopeTeam          = "team_manager"
)

type ManagerContext struct {
	EmployeeID            uuid.UUID
	ManagementScope       string
	ManagedDepartmentIDs  []uuid.UUID
	ManagedDepartmentRoots []uuid.UUID
	ManagedTeamIDs        []uuid.UUID
	ManagedTeamMemberIDs  []uuid.UUID
}

func ParseUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse uuid: %w", err)
	}
	return value, nil
}

func CurrentUserUUID(claims *internalauth.AccessClaims) (*uuid.UUID, error) {
	if claims == nil || claims.UserID == "" {
		return nil, nil
	}
	value, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func CurrentEmployeeUUID(claims *internalauth.AccessClaims) (*uuid.UUID, error) {
	if claims == nil || claims.EmployeeID == "" {
		return nil, nil
	}
	value, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func FullName(firstName, lastName string) string {
	return strings.TrimSpace(firstName + " " + lastName)
}

func FindOrganization(db *gorm.DB) (models.Organization, error) {
	var organization models.Organization
	err := db.Where("code = ?", DefaultOrganizationCode).First(&organization).Error
	return organization, err
}

func MustOrganizationID(db *gorm.DB) (uuid.UUID, error) {
	organization, err := FindOrganization(db)
	if err != nil {
		return uuid.Nil, err
	}
	return organization.ID, nil
}

func PreloadEmployeeDetails(query *gorm.DB) *gorm.DB {
	return query.
		Preload("Department").
		Preload("Job.JobLevel").
		Preload("Job.PrimaryDepartment").
		Preload("Location").
		Preload("EmployeeType").
		Preload("EmploymentStatus").
		Preload("Manager").
		Preload("ManagedTeams.Department").
		Preload("TeamMemberships.Team").
		Preload("RoleAssignments.Job.JobLevel").
		Preload("RoleAssignments.Department")
}

func FindLinkedEmployee(db *gorm.DB, user models.User) (*models.Employee, error) {
	query := PreloadEmployeeDetails(db)
	if user.EmployeeID != nil {
		var employee models.Employee
		result := query.Limit(1).Find(&employee, "id = ?", *user.EmployeeID)
		if result.Error == nil && result.RowsAffected > 0 {
			return &employee, nil
		}
	}

	var employee models.Employee
	result := query.Limit(1).Find(&employee, "user_id = ?", user.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &employee, nil
}

func ManagedDepartmentRootIDs(db *gorm.DB, managerEmployeeID uuid.UUID) ([]uuid.UUID, error) {
	type row struct {
		ID uuid.UUID
	}
	var rows []row
	if err := db.Raw(`
SELECT id
FROM departments
WHERE manager_employee_id = ?
ORDER BY level ASC, name ASC
`, managerEmployeeID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, item := range rows {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func DepartmentTreeIDs(db *gorm.DB, rootDepartmentIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(rootDepartmentIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	seen := map[uuid.UUID]bool{}
	queue := append([]uuid.UUID{}, rootDepartmentIDs...)
	result := make([]uuid.UUID, 0, len(rootDepartmentIDs))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		result = append(result, current)

		var children []uuid.UUID
		if err := db.Model(&models.Department{}).Where("parent_department_id = ?", current).Pluck("id", &children).Error; err != nil {
			return nil, err
		}
		queue = append(queue, children...)
	}

	return result, nil
}

func AccessibleDepartmentIDs(db *gorm.DB, managerEmployeeID uuid.UUID) ([]uuid.UUID, error) {
	roots, err := ManagedDepartmentRootIDs(db, managerEmployeeID)
	if err != nil {
		return nil, err
	}
	return DepartmentTreeIDs(db, roots)
}

func ManagedTeamIDs(db *gorm.DB, managerEmployeeID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := db.Model(&models.Team{}).
		Where("manager_employee_id = ?", managerEmployeeID).
		Order("name asc").
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func ActiveTeamMemberIDs(db *gorm.DB, teamIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(teamIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var ids []uuid.UUID
	if err := db.Model(&models.TeamMembership{}).
		Where("team_id IN ? AND (end_date IS NULL OR end_date >= CURRENT_DATE)", teamIDs).
		Distinct("employee_id").
		Pluck("employee_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func LoadManagerContext(db *gorm.DB, managerEmployeeID uuid.UUID) (ManagerContext, error) {
	var employee models.Employee
	if err := db.Select("id", "management_scope").First(&employee, "id = ?", managerEmployeeID).Error; err != nil {
		return ManagerContext{}, err
	}

	departmentRoots, err := ManagedDepartmentRootIDs(db, managerEmployeeID)
	if err != nil {
		return ManagerContext{}, err
	}
	departmentIDs, err := DepartmentTreeIDs(db, departmentRoots)
	if err != nil {
		return ManagerContext{}, err
	}
	teamIDs, err := ManagedTeamIDs(db, managerEmployeeID)
	if err != nil {
		return ManagerContext{}, err
	}
	teamMemberIDs, err := ActiveTeamMemberIDs(db, teamIDs)
	if err != nil {
		return ManagerContext{}, err
	}

	return ManagerContext{
		EmployeeID:             managerEmployeeID,
		ManagementScope:        employee.ManagementScope,
		ManagedDepartmentIDs:   departmentIDs,
		ManagedDepartmentRoots: departmentRoots,
		ManagedTeamIDs:         teamIDs,
		ManagedTeamMemberIDs:   teamMemberIDs,
	}, nil
}

func EmployeeWithinManagerScope(employee models.Employee, managerID uuid.UUID, departmentIDs []uuid.UUID) bool {
	if employee.ManagerEmployeeID != nil && *employee.ManagerEmployeeID == managerID {
		return true
	}
	for _, id := range departmentIDs {
		if employee.DepartmentID == id {
			return true
		}
	}
	return false
}

func EmployeeWithinManagerContext(employee models.Employee, context ManagerContext) bool {
	switch context.ManagementScope {
	case ManagementScopeDivision, ManagementScopeSubdepartment:
		for _, id := range context.ManagedDepartmentIDs {
			if employee.DepartmentID == id {
				return true
			}
		}
		return false
	case ManagementScopeTeam:
		for _, id := range context.ManagedTeamMemberIDs {
			if employee.ID == id {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func FindNearestParentManagerID(db *gorm.DB, employeeID uuid.UUID) (*uuid.UUID, error) {
	type employeeRow struct {
		DepartmentID uuid.UUID
	}

	var employee employeeRow
	if err := db.Table("employees").Select("department_id").First(&employee, "id = ?", employeeID).Error; err != nil {
		return nil, err
	}

	type departmentRow struct {
		ID                uuid.UUID
		ParentDepartmentID *uuid.UUID
		ManagerEmployeeID *uuid.UUID
	}

	var current departmentRow
	if err := db.Table("departments").Select("id", "parent_department_id", "manager_employee_id").First(&current, "id = ?", employee.DepartmentID).Error; err != nil {
		return nil, err
	}

	visited := map[uuid.UUID]bool{}
	parentID := current.ParentDepartmentID
	for parentID != nil {
		if visited[*parentID] {
			break
		}
		visited[*parentID] = true

		var parent departmentRow
		if err := db.Table("departments").Select("id", "parent_department_id", "manager_employee_id").First(&parent, "id = ?", *parentID).Error; err != nil {
			return nil, err
		}
		if parent.ManagerEmployeeID != nil && *parent.ManagerEmployeeID != employeeID {
			return parent.ManagerEmployeeID, nil
		}
		parentID = parent.ParentDepartmentID
	}

	return nil, nil
}

func FindAdminApproverEmployeeID(db *gorm.DB, organizationID uuid.UUID) (*uuid.UUID, error) {
	type row struct {
		EmployeeID *uuid.UUID
	}

	var result row
	err := db.Table("users").
		Select("users.employee_id").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("users.organization_id = ? AND roles.code = ? AND users.employee_id IS NOT NULL", organizationID, "admin").
		Limit(1).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return result.EmployeeID, nil
}

func RequiresRemoteAttendance(workMode, employeeTypeCode string) bool {
	workMode = strings.ToLower(strings.TrimSpace(workMode))
	employeeTypeCode = strings.ToLower(strings.TrimSpace(employeeTypeCode))
	return workMode == "remote" || workMode == "client-based" || employeeTypeCode == "freelance"
}

func AssignmentStatus(startDate time.Time, endDate *time.Time, now time.Time) string {
	currentDay := now.UTC().Truncate(24 * time.Hour)
	startDay := startDate.UTC().Truncate(24 * time.Hour)
	if endDate != nil {
		endDay := endDate.UTC().Truncate(24 * time.Hour)
		if endDay.Before(currentDay) {
			return "ended"
		}
	}
	if startDay.After(currentDay) {
		return "scheduled"
	}
	return "active"
}
