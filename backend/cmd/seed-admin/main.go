package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hrms/backend/internal/auth"
	"hrms/backend/internal/config"
	"hrms/backend/internal/db"
	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/shared"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.AdminSeedEmail == "" || cfg.AdminSeedPassword == "" {
		log.Fatal("ADMIN_SEED_EMAIL and ADMIN_SEED_PASSWORD are required")
	}

	gormDB, err := db.OpenGORM(cfg)
	if err != nil {
		log.Fatal(err)
	}

	hasher := auth.NewPasswordHasher()
	hash, err := hasher.Hash(cfg.AdminSeedPassword)
	if err != nil {
		log.Fatal(err)
	}

	var role models.Role
	if err := gormDB.Where("code = ?", "admin").First(&role).Error; err != nil {
		log.Fatal(err)
	}

	var organization models.Organization
	err = gormDB.Where("code = ?", shared.DefaultOrganizationCode).
		FirstOrCreate(&organization, models.Organization{Code: shared.DefaultOrganizationCode, Name: "PT. CODEID"}).Error
	if err != nil {
		log.Fatal(err)
	}

	email := strings.ToLower(cfg.AdminSeedEmail)
	user := models.User{}
	assignments := map[string]any{
		"organization_id": organization.ID,
		"password_hash":   hash,
		"role_id":         role.ID,
		"is_active":       true,
		"last_login_at":   ptrTime(time.Now().UTC()),
	}
	if err := gormDB.Where("email = ?", email).
		Assign(assignments).
		Attrs(models.User{
			ID:             uuid.New(),
			OrganizationID: organization.ID,
			Email:          email,
			PasswordHash:   hash,
			RoleID:         role.ID,
			IsActive:       true,
			LastLoginAt:    ptrTime(time.Now().UTC()),
		}).
		FirstOrCreate(&user).Error; err != nil {
		log.Fatal(err)
	}

	if err := ensureAdminEmployee(gormDB, organization.ID, &user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("admin user %s ensured, but reference data is not complete yet for a linked employee; run seed-demo after reference setup\n", user.Email)
			return
		}
		log.Fatal(err)
	}

	log.Printf("seeded admin user %s with linked employee context\n", user.Email)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func ensureAdminEmployee(gormDB *gorm.DB, organizationID uuid.UUID, user *models.User) error {
	var employeeType models.EmployeeType
	if err := gormDB.Where("organization_id = ? AND code = ?", organizationID, "full-time").First(&employeeType).Error; err != nil {
		return err
	}

	var employmentStatus models.EmploymentStatus
	if err := gormDB.Where("organization_id = ? AND code = ?", organizationID, "active").First(&employmentStatus).Error; err != nil {
		return err
	}

	var department models.Department
	if err := gormDB.Where("organization_id = ? AND name = ?", organizationID, "Divisi HR").First(&department).Error; err != nil {
		return err
	}

	var job models.Job
	if err := gormDB.Where("organization_id = ? AND title = ?", organizationID, "HR Manager").First(&job).Error; err != nil {
		return err
	}

	var location models.Location
	if err := gormDB.Where("organization_id = ? AND name = ?", organizationID, "Jakarta Headquarters").First(&location).Error; err != nil {
		return err
	}

	var employee models.Employee
	hireDate := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	assignments := map[string]any{
		"organization_id":        organizationID,
		"user_id":                user.ID,
		"first_name":             "System",
		"last_name":              "Administrator",
		"email":                  user.Email,
		"phone_number":           "+62-21-0000-0000",
		"hire_date":              hireDate,
		"employee_type_id":       employeeType.ID,
		"employment_status_id":   employmentStatus.ID,
		"department_id":          department.ID,
		"job_id":                 job.ID,
		"location_id":            location.ID,
		"work_mode":              "onsite",
		"management_scope":       shared.ManagementScopeIndividual,
		"manager_employee_id":    nil,
	}
	if err := gormDB.Where("user_id = ? OR employee_code = ?", user.ID, "CID-ADMIN-001").
		Assign(assignments).
		Attrs(models.Employee{
			ID:                 uuid.New(),
			OrganizationID:     organizationID,
			EmployeeCode:       "CID-ADMIN-001",
			UserID:             &user.ID,
			FirstName:          "System",
			LastName:           "Administrator",
			Email:              user.Email,
			PhoneNumber:        "+62-21-0000-0000",
			HireDate:           hireDate,
			EmployeeTypeID:     employeeType.ID,
			EmploymentStatusID: employmentStatus.ID,
			DepartmentID:       department.ID,
			JobID:              job.ID,
			LocationID:         location.ID,
			WorkMode:           "onsite",
			ManagementScope:    shared.ManagementScopeIndividual,
		}).
		FirstOrCreate(&employee).Error; err != nil {
		return err
	}

	return gormDB.Model(&models.User{}).Where("id = ?", user.ID).Update("employee_id", employee.ID).Error
}
