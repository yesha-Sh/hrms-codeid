package lookups

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	internalauth "hrms/backend/internal/auth"
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
	router.Route("/lookups", func(r chi.Router) {
		r.Get("/employees", service.employees)
		r.Get("/departments", service.departments)
		r.Get("/jobs", service.jobs)
		r.Get("/job-levels", service.jobLevels)
		r.Get("/locations", service.locations)
		r.Get("/employee-types", service.employeeTypes)
		r.Get("/leave-types", service.leaveTypes)
		r.Get("/employment-statuses", service.employmentStatuses)
		r.Get("/holidays", service.holidays)
	})
}

func (s Service) employees(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	query := s.db.Model(&models.Employee{}).
		Where("employees.organization_id = ?", organizationID).
		Preload("Department").
		Preload("Job")

	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		term := "%" + search + "%"
		query = query.Where(
			"employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR employees.employee_code ILIKE ? OR employees.email ILIKE ?",
			term,
			term,
			term,
			term,
		)
	}
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("employees.department_id = ?", departmentID)
	}

	scoped, err := s.applyEmployeeScope(query, claims)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	var employees []models.Employee
	if err := scoped.Order("employees.first_name asc, employees.last_name asc").Find(&employees).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load employee lookup", nil)
		return
	}

	items := make([]map[string]any, 0, len(employees))
	for _, employee := range employees {
		metaParts := []string{employee.EmployeeCode, employee.Department.Name, employee.Job.Title}
		items = append(items, map[string]any{
			"id":    employee.ID,
			"label": shared.FullName(employee.FirstName, employee.LastName),
			"meta":  strings.Join(compact(metaParts), " · "),
			"context": map[string]any{
				"department_id":     employee.DepartmentID,
				"work_mode":         employee.WorkMode,
				"management_scope":  employee.ManagementScope,
			},
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s Service) departments(w http.ResponseWriter, wReq *http.Request) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	query := s.db.Model(&models.Department{}).
		Where("organization_id = ?", organizationID).
		Preload("Location").
		Preload("ParentDepartment").
		Order("level asc, name asc")

	if parentID := wReq.URL.Query().Get("parent_department_id"); parentID != "" {
		query = query.Where("parent_department_id = ?", parentID)
	}
	if level := wReq.URL.Query().Get("level"); level != "" {
		query = query.Where("level = ?", level)
	}

	var departments []models.Department
	if err := query.Find(&departments).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load department lookup", nil)
		return
	}

	items := make([]map[string]any, 0, len(departments))
	for _, department := range departments {
		metaParts := []string{fmt.Sprintf("Level %d", department.Level), department.Location.Name}
		if department.ParentDepartment != nil {
			metaParts = append(metaParts, "Parent: "+department.ParentDepartment.Name)
		}
		items = append(items, map[string]any{
			"id":    department.ID,
			"label": department.Name,
			"meta":  strings.Join(compact(metaParts), " · "),
			"context": map[string]any{
				"level":                department.Level,
				"parent_department_id": department.ParentDepartmentID,
			},
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s Service) jobs(w http.ResponseWriter, r *http.Request) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	query := s.db.Model(&models.Job{}).
		Where("organization_id = ?", organizationID).
		Preload("PrimaryDepartment").
		Preload("JobLevel").
		Order("title asc")

	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("primary_department_id = ?", departmentID)
	}
	if levelID := r.URL.Query().Get("job_level_id"); levelID != "" {
		query = query.Where("job_level_id = ?", levelID)
	}

	var jobs []models.Job
	if err := query.Find(&jobs).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load job lookup", nil)
		return
	}

	items := make([]map[string]any, 0, len(jobs))
	for _, job := range jobs {
		metaParts := []string{job.PrimaryDepartment.Name, job.JobLevel.Name, salaryLabel(job.MinSalary, job.MaxSalary)}
		items = append(items, map[string]any{
			"id":    job.ID,
			"label": job.Title,
			"meta":  strings.Join(compact(metaParts), " · "),
			"context": map[string]any{
				"primary_department_id": job.PrimaryDepartmentID,
				"job_level_id":          job.JobLevelID,
			},
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s Service) jobLevels(w http.ResponseWriter, r *http.Request) {
	s.referenceLookup(w, r, &[]models.JobLevel{}, "job_levels", "name asc", func(item any) map[string]any {
		level := item.(models.JobLevel)
		return map[string]any{"id": level.ID, "label": level.Name, "meta": fmt.Sprintf("Rank %d", level.Rank)}
	})
}

func (s Service) locations(w http.ResponseWriter, r *http.Request) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	var locations []models.Location
	if err := s.db.Model(&models.Location{}).
		Where("organization_id = ?", organizationID).
		Order("name asc").
		Find(&locations).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load location lookup", nil)
		return
	}

	items := make([]map[string]any, 0, len(locations))
	for _, location := range locations {
		metaParts := []string{location.City, location.Address}
		items = append(items, map[string]any{
			"id":    location.ID,
			"label": location.Name,
			"meta":  strings.Join(compact(metaParts), " · "),
			"context": map[string]any{
				"city": location.City,
			},
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s Service) employeeTypes(w http.ResponseWriter, r *http.Request) {
	s.referenceLookup(w, r, &[]models.EmployeeType{}, "employee_types", "name asc", func(item any) map[string]any {
		reference := item.(models.EmployeeType)
		return map[string]any{"id": reference.ID, "label": reference.Name, "meta": reference.Code}
	})
}

func (s Service) leaveTypes(w http.ResponseWriter, r *http.Request) {
	s.referenceLookup(w, r, &[]models.LeaveType{}, "leave_types", "name asc", func(item any) map[string]any {
		reference := item.(models.LeaveType)
		metaParts := []string{reference.Code}
		if reference.IsPaid {
			metaParts = append(metaParts, "Paid")
		} else {
			metaParts = append(metaParts, "Unpaid")
		}
		return map[string]any{"id": reference.ID, "label": reference.Name, "meta": strings.Join(metaParts, " · ")}
	})
}

func (s Service) employmentStatuses(w http.ResponseWriter, r *http.Request) {
	s.referenceLookup(w, r, &[]models.EmploymentStatus{}, "employment_statuses", "name asc", func(item any) map[string]any {
		reference := item.(models.EmploymentStatus)
		return map[string]any{"id": reference.ID, "label": reference.Name, "meta": reference.Code}
	})
}

func (s Service) holidays(w http.ResponseWriter, r *http.Request) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	query := s.db.Model(&models.Holiday{}).
		Where("organization_id = ?", organizationID).
		Preload("Location").
		Order("holiday_date asc")
	if year := r.URL.Query().Get("year"); year != "" {
		query = query.Where("EXTRACT(YEAR FROM holiday_date) = ?", year)
	}

	var holidays []models.Holiday
	if err := query.Find(&holidays).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load holiday lookup", nil)
		return
	}

	items := make([]map[string]any, 0, len(holidays))
	for _, holiday := range holidays {
		metaParts := []string{holiday.HolidayDate.Format("2006-01-02")}
		if holiday.Location != nil {
			metaParts = append(metaParts, holiday.Location.Name)
		}
		items = append(items, map[string]any{
			"id":    holiday.ID,
			"label": holiday.Name,
			"meta":  strings.Join(metaParts, " · "),
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s Service) referenceLookup(w http.ResponseWriter, r *http.Request, target any, table string, order string, mapItem func(any) map[string]any) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	switch items := target.(type) {
	case *[]models.JobLevel:
		if err := s.db.Table(table).Where("organization_id = ?", organizationID).Order(order).Scan(items).Error; err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not load lookup", nil)
			return
		}
		response := make([]map[string]any, 0, len(*items))
		for _, item := range *items {
			response = append(response, mapItem(item))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"items": response})
	case *[]models.EmployeeType:
		if err := s.db.Table(table).Where("organization_id = ?", organizationID).Order(order).Scan(items).Error; err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not load lookup", nil)
			return
		}
		response := make([]map[string]any, 0, len(*items))
		for _, item := range *items {
			response = append(response, mapItem(item))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"items": response})
	case *[]models.LeaveType:
		if err := s.db.Table(table).Where("organization_id = ?", organizationID).Order(order).Scan(items).Error; err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not load lookup", nil)
			return
		}
		response := make([]map[string]any, 0, len(*items))
		for _, item := range *items {
			response = append(response, mapItem(item))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"items": response})
	case *[]models.EmploymentStatus:
		if err := s.db.Table(table).Where("organization_id = ?", organizationID).Order(order).Scan(items).Error; err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not load lookup", nil)
			return
		}
		response := make([]map[string]any, 0, len(*items))
		for _, item := range *items {
			response = append(response, mapItem(item))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"items": response})
	default:
		httpx.Error(w, http.StatusInternalServerError, "unsupported lookup target", nil)
	}
}

func (s Service) applyEmployeeScope(query *gorm.DB, claims *internalauth.AccessClaims) (*gorm.DB, error) {
	if claims == nil {
		return nil, errors.New("missing auth claims")
	}

	switch claims.Role {
	case "admin":
		return query, nil
	case "manager":
		employeeID, err := shared.CurrentEmployeeUUID(claims)
		if err != nil || employeeID == nil {
			return nil, errors.New("manager employee context is required")
		}
		context, err := shared.LoadManagerContext(s.db, *employeeID)
		if err != nil {
			return nil, err
		}
		switch context.ManagementScope {
		case shared.ManagementScopeDivision, shared.ManagementScopeSubdepartment:
			if len(context.ManagedDepartmentIDs) == 0 {
				return query.Where("employees.id = ?", *employeeID), nil
			}
			return query.Where("employees.id = ? OR employees.department_id IN ?", *employeeID, context.ManagedDepartmentIDs), nil
		case shared.ManagementScopeTeam:
			ids := append([]uuid.UUID{*employeeID}, context.ManagedTeamMemberIDs...)
			return query.Where("employees.id IN ?", ids), nil
		default:
			return query.Where("employees.id = ?", *employeeID), nil
		}
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

func salaryLabel(min *float64, max *float64) string {
	if min == nil && max == nil {
		return "Salary not set"
	}
	if min == nil {
		return "Up to configured max"
	}
	if max == nil {
		return "From configured min"
	}
	return fmt.Sprintf("%.2f - %.2f", *min, *max)
}

func compact(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
