package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hrms/backend/internal/auth"
	"hrms/backend/internal/config"
	"hrms/backend/internal/db"
	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/shared"
)

type departmentSeed struct {
	Name       string
	ParentName string
	Location   string
	Level      int
}

type jobSeed struct {
	Title       string
	Department  string
	Level       string
	MinSalary   float64
	MaxSalary   float64
	Description string
}

type employeeSeed struct {
	Role             string
	Email            string
	Password         string
	EmployeeCode     string
	FirstName        string
	LastName         string
	Phone            string
	Department       string
	Job              string
	Location         string
	ManagerEmail     string
	EmployeeType     string
	EmploymentStatus string
	WorkMode         string
	ManagementScope  string
	Salary           float64
	HireDate         string
	Secondary        []secondarySeed
}

type secondarySeed struct {
	Job          string
	Department   string
	HoursPerWeek float64
	StartDate    string
	EndDate      string
	Notes        string
}

type teamSeed struct {
	Name              string
	Department        string
	ManagerEmail      string
	IsCrossDepartment bool
	FocusArea         string
	Members           []string
}

var employeeSeeds = []employeeSeed{
	{Role: "manager", Email: "aria.pratama@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-001", FirstName: "Aria", LastName: "Pratama", Phone: "+62-811-1001", Department: "Divisi IT", Job: "IT Manager", Location: "Jakarta Headquarters", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeDivision, Salary: 24000000, HireDate: "2021-03-15"},
	{Role: "manager", Email: "sinta.lestari@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-002", FirstName: "Sinta", LastName: "Lestari", Phone: "+62-811-1002", Department: "Divisi Business & Development", Job: "Business Development Manager", Location: "Jakarta Headquarters", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeDivision, Salary: 25000000, HireDate: "2020-09-01"},
	{Role: "manager", Email: "bayu.saputra@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-003", FirstName: "Bayu", LastName: "Saputra", Phone: "+62-811-1003", Department: "Divisi Sales", Job: "Sales Manager", Location: "Jakarta Headquarters", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeDivision, Salary: 21000000, HireDate: "2021-05-10"},
	{Role: "manager", Email: "rina.wulandari@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-004", FirstName: "Rina", LastName: "Wulandari", Phone: "+62-811-1004", Department: "Divisi Finance", Job: "Finance Manager", Location: "Jakarta Headquarters", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeDivision, Salary: 22500000, HireDate: "2019-11-18"},
	{Role: "manager", Email: "dimas.kusuma@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-005", FirstName: "Dimas", LastName: "Kusuma", Phone: "+62-811-1005", Department: "Divisi HR", Job: "HR Manager", Location: "Jakarta Headquarters", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeDivision, Salary: 20500000, HireDate: "2020-01-08"},
	{Role: "manager", Email: "hanif.ardana@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-022", FirstName: "Hanif", LastName: "Ardana", Phone: "+62-811-1022", Department: "Infrastructure", Job: "Infrastructure Lead", Location: "Jakarta Headquarters", ManagerEmail: "aria.pratama@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeSubdepartment, Salary: 19250000, HireDate: "2022-03-07"},
	{Role: "manager", Email: "nadia.hapsari@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-023", FirstName: "Nadia", LastName: "Hapsari", Phone: "+62-811-1023", Department: "Smartsourcing", Job: "Smartsourcing Lead", Location: "Remote / Client Location", ManagerEmail: "sinta.lestari@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeSubdepartment, Salary: 18800000, HireDate: "2022-05-16"},
	{Role: "manager", Email: "salma.nuraini@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-011", FirstName: "Salma", LastName: "Nuraini", Phone: "+62-811-1011", Department: "System Development", Job: "System Development Lead", Location: "Jakarta Headquarters", ManagerEmail: "aria.pratama@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeSubdepartment, Salary: 19800000, HireDate: "2022-02-14"},
	{Role: "manager", Email: "luthfi.ramadhan@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-012", FirstName: "Luthfi", LastName: "Ramadhan", Phone: "+62-811-1012", Department: "Bootcamp", Job: "Bootcamp Lead", Location: "Bandung Branch", ManagerEmail: "sinta.lestari@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeSubdepartment, Salary: 18250000, HireDate: "2022-06-21"},
	{Role: "manager", Email: "fajar.maulana@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-013", FirstName: "Fajar", LastName: "Maulana", Phone: "+62-811-1013", Department: "System Development", Job: "System Development Lead", Location: "Jakarta Headquarters", ManagerEmail: "salma.nuraini@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeTeam, Salary: 17500000, HireDate: "2023-03-01"},
	{Role: "manager", Email: "mira.anggraeni@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-014", FirstName: "Mira", LastName: "Anggraeni", Phone: "+62-811-1014", Department: "Bootcamp", Job: "Bootcamp Lead", Location: "Bandung Branch", ManagerEmail: "luthfi.ramadhan@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeTeam, Salary: 16800000, HireDate: "2023-08-07"},
	{Role: "manager", Email: "dewi.pangestu@codeid.co.id", Password: "Manager123!", EmployeeCode: "CID-024", FirstName: "Dewi", LastName: "Pangestu", Phone: "+62-811-1024", Department: "Divisi Sales", Job: "Sales Team Lead", Location: "Jakarta Headquarters", ManagerEmail: "bayu.saputra@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeTeam, Salary: 16300000, HireDate: "2023-10-09"},
	{Role: "employee", Email: "nabila.putri@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-006", FirstName: "Nabila", LastName: "Putri", Phone: "+62-811-1006", Department: "System Development", Job: "System Developer", Location: "Jakarta Headquarters", ManagerEmail: "salma.nuraini@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 16500000, HireDate: "2022-07-12", Secondary: []secondarySeed{{Job: "Outsourcing Specialist", Department: "Smartsourcing", HoursPerWeek: 8, StartDate: "2026-01-15", Notes: "Short-term client delivery support assignment."}}},
	{Role: "employee", Email: "raka.pradipta@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-015", FirstName: "Raka", LastName: "Pradipta", Phone: "+62-811-1015", Department: "System Development", Job: "System Developer", Location: "Jakarta Headquarters", ManagerEmail: "fajar.maulana@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 15250000, HireDate: "2024-01-15"},
	{Role: "employee", Email: "dian.permata@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-016", FirstName: "Dian", LastName: "Permata", Phone: "+62-811-1016", Department: "Infrastructure", Job: "Network Engineer", Location: "Jakarta Headquarters", ManagerEmail: "aria.pratama@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 14250000, HireDate: "2023-11-06"},
	{Role: "employee", Email: "farhan.akbar@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-007", FirstName: "Farhan", LastName: "Akbar", Phone: "+62-811-1007", Department: "Security", Job: "Security Specialist", Location: "Jakarta Headquarters", ManagerEmail: "aria.pratama@codeid.co.id", EmployeeType: "contract", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 17250000, HireDate: "2023-02-03"},
	{Role: "employee", Email: "keisha.anindya@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-008", FirstName: "Keisha", LastName: "Anindya", Phone: "+62-811-1008", Department: "Bootcamp", Job: "Talent Acquisition Specialist", Location: "Bandung Branch", ManagerEmail: "luthfi.ramadhan@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 13500000, HireDate: "2022-10-20", Secondary: []secondarySeed{{Job: "Recruitment Specialist", Department: "Divisi HR", HoursPerWeek: 6, StartDate: "2026-02-03", Notes: "Supports shared hiring funnel for bootcamp instructors."}}},
	{Role: "employee", Email: "yoga.prasetyo@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-017", FirstName: "Yoga", LastName: "Prasetyo", Phone: "+62-811-1017", Department: "Divisi HR", Job: "Recruitment Specialist", Location: "Jakarta Headquarters", ManagerEmail: "dimas.kusuma@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 11800000, HireDate: "2023-09-18", Secondary: []secondarySeed{{Job: "Talent Acquisition Specialist", Department: "Bootcamp", HoursPerWeek: 10, StartDate: "2026-01-20", Notes: "Supports bootcamp intake and facilitator sourcing."}}},
	{Role: "employee", Email: "vina.mahendra@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-018", FirstName: "Vina", LastName: "Mahendra", Phone: "+62-811-1018", Department: "Divisi HR", Job: "Payroll Specialist", Location: "Jakarta Headquarters", ManagerEmail: "dimas.kusuma@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 12250000, HireDate: "2022-12-01"},
	{Role: "employee", Email: "talia.rahma@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-009", FirstName: "Talia", LastName: "Rahma", Phone: "+62-811-1009", Department: "Smartsourcing", Job: "Outsourcing Specialist", Location: "Remote / Client Location", ManagerEmail: "sinta.lestari@codeid.co.id", EmployeeType: "freelance", EmploymentStatus: "active", WorkMode: "remote", ManagementScope: shared.ManagementScopeIndividual, Salary: 12800000, HireDate: "2023-05-14"},
	{Role: "employee", Email: "aulia.kirana@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-019", FirstName: "Aulia", LastName: "Kirana", Phone: "+62-811-1019", Department: "Smartsourcing", Job: "Outsourcing Specialist", Location: "Remote / Client Location", ManagerEmail: "sinta.lestari@codeid.co.id", EmployeeType: "contract", EmploymentStatus: "active", WorkMode: "client-based", ManagementScope: shared.ManagementScopeIndividual, Salary: 13250000, HireDate: "2024-02-12"},
	{Role: "employee", Email: "rafi.hakim@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-010", FirstName: "Rafi", LastName: "Hakim", Phone: "+62-811-1010", Department: "Divisi Sales", Job: "Account Executive", Location: "Jakarta Headquarters", ManagerEmail: "bayu.saputra@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "probation", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 11000000, HireDate: "2026-01-10"},
	{Role: "employee", Email: "gerry.wijaya@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-020", FirstName: "Gerry", LastName: "Wijaya", Phone: "+62-811-1020", Department: "Divisi Sales", Job: "Sales Consultant", Location: "Jakarta Headquarters", ManagerEmail: "bayu.saputra@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 9200000, HireDate: "2024-04-22"},
	{Role: "employee", Email: "putri.ardelia@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-021", FirstName: "Putri", LastName: "Ardelia", Phone: "+62-811-1021", Department: "Divisi Finance", Job: "Accountant", Location: "Jakarta Headquarters", ManagerEmail: "rina.wulandari@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 12600000, HireDate: "2023-06-12"},
	{Role: "employee", Email: "bima.satrya@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-025", FirstName: "Bima", LastName: "Satrya", Phone: "+62-811-1025", Department: "System Development", Job: "Senior System Developer", Location: "Jakarta Headquarters", ManagerEmail: "salma.nuraini@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 18400000, HireDate: "2021-08-23", Secondary: []secondarySeed{{Job: "Bootcamp Program Coordinator", Department: "Bootcamp", HoursPerWeek: 6, StartDate: "2026-02-10", Notes: "Technical mentor for the April bootcamp cohort."}}},
	{Role: "employee", Email: "intan.maharani@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-026", FirstName: "Intan", LastName: "Maharani", Phone: "+62-811-1026", Department: "System Development", Job: "Junior System Developer", Location: "Jakarta Headquarters", ManagerEmail: "fajar.maulana@codeid.co.id", EmployeeType: "internship", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 6500000, HireDate: "2026-02-02"},
	{Role: "employee", Email: "claudia.suryani@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-027", FirstName: "Claudia", LastName: "Suryani", Phone: "+62-811-1027", Department: "Infrastructure", Job: "Network Engineer", Location: "Jakarta Headquarters", ManagerEmail: "hanif.ardana@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 13600000, HireDate: "2024-08-05"},
	{Role: "employee", Email: "ryan.hadikusuma@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-028", FirstName: "Ryan", LastName: "Hadikusuma", Phone: "+62-811-1028", Department: "Security", Job: "Security Specialist", Location: "Jakarta Headquarters", ManagerEmail: "aria.pratama@codeid.co.id", EmployeeType: "contract", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 15800000, HireDate: "2025-01-13"},
	{Role: "employee", Email: "rio.firmansyah@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-029", FirstName: "Rio", LastName: "Firmansyah", Phone: "+62-811-1029", Department: "Smartsourcing", Job: "Outsourcing Specialist", Location: "Remote / Client Location", ManagerEmail: "nadia.hapsari@codeid.co.id", EmployeeType: "contract", EmploymentStatus: "active", WorkMode: "client-based", ManagementScope: shared.ManagementScopeIndividual, Salary: 13900000, HireDate: "2024-06-03", Secondary: []secondarySeed{{Job: "Account Executive", Department: "Divisi Sales", HoursPerWeek: 4, StartDate: "2026-03-01", Notes: "Supports handoff for strategic staffing accounts."}}},
	{Role: "employee", Email: "laras.ayu@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-030", FirstName: "Laras", LastName: "Ayu", Phone: "+62-811-1030", Department: "Smartsourcing", Job: "Client Success Analyst", Location: "Remote / Client Location", ManagerEmail: "nadia.hapsari@codeid.co.id", EmployeeType: "freelance", EmploymentStatus: "active", WorkMode: "remote", ManagementScope: shared.ManagementScopeIndividual, Salary: 10800000, HireDate: "2025-07-07"},
	{Role: "employee", Email: "ajeng.kusnadi@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-031", FirstName: "Ajeng", LastName: "Kusnadi", Phone: "+62-811-1031", Department: "Bootcamp", Job: "Bootcamp Program Coordinator", Location: "Bandung Branch", ManagerEmail: "mira.anggraeni@codeid.co.id", EmployeeType: "part-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 9200000, HireDate: "2025-03-17"},
	{Role: "employee", Email: "mei.lestari@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-032", FirstName: "Mei", LastName: "Lestari", Phone: "+62-811-1032", Department: "Divisi HR", Job: "People Operations Specialist", Location: "Jakarta Headquarters", ManagerEmail: "dimas.kusuma@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "on-leave", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 12400000, HireDate: "2024-09-02"},
	{Role: "employee", Email: "bagas.nugroho@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-033", FirstName: "Bagas", LastName: "Nugroho", Phone: "+62-811-1033", Department: "Divisi Finance", Job: "Finance Analyst", Location: "Jakarta Headquarters", ManagerEmail: "rina.wulandari@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 9800000, HireDate: "2025-05-12"},
	{Role: "employee", Email: "selvi.anggraini@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-034", FirstName: "Selvi", LastName: "Anggraini", Phone: "+62-811-1034", Department: "Divisi Sales", Job: "Account Executive", Location: "Jakarta Headquarters", ManagerEmail: "dewi.pangestu@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "active", WorkMode: "hybrid", ManagementScope: shared.ManagementScopeIndividual, Salary: 11600000, HireDate: "2025-02-10"},
	{Role: "employee", Email: "nico.febriansyah@codeid.co.id", Password: "Employee123!", EmployeeCode: "CID-035", FirstName: "Nico", LastName: "Febriansyah", Phone: "+62-811-1035", Department: "Divisi Sales", Job: "Sales Consultant", Location: "Jakarta Headquarters", ManagerEmail: "dewi.pangestu@codeid.co.id", EmployeeType: "full-time", EmploymentStatus: "probation", WorkMode: "onsite", ManagementScope: shared.ManagementScopeIndividual, Salary: 8900000, HireDate: "2026-02-16"},
}

var teamSeeds = []teamSeed{
	{Name: "Platform Delivery Team", Department: "System Development", ManagerEmail: "fajar.maulana@codeid.co.id", IsCrossDepartment: false, FocusArea: "Owns day-to-day delivery for platform backlog and release readiness.", Members: []string{"nabila.putri@codeid.co.id", "raka.pradipta@codeid.co.id"}},
	{Name: "Client Solutions Squad", Department: "System Development", ManagerEmail: "fajar.maulana@codeid.co.id", IsCrossDepartment: true, FocusArea: "Cross-department delivery squad for client solutions and shared project execution.", Members: []string{"talia.rahma@codeid.co.id", "rafi.hakim@codeid.co.id", "gerry.wijaya@codeid.co.id"}},
	{Name: "Learning Operations Pod", Department: "Bootcamp", ManagerEmail: "mira.anggraeni@codeid.co.id", IsCrossDepartment: true, FocusArea: "Coordinates bootcamp delivery, hiring readiness, and participant operations across supporting functions.", Members: []string{"keisha.anindya@codeid.co.id", "yoga.prasetyo@codeid.co.id"}},
	{Name: "Revenue Acceleration Pod", Department: "Divisi Sales", ManagerEmail: "dewi.pangestu@codeid.co.id", IsCrossDepartment: true, FocusArea: "Connects sales execution, client staffing readiness, and cross-functional delivery handoff for new accounts.", Members: []string{"selvi.anggraini@codeid.co.id", "nico.febriansyah@codeid.co.id", "rio.firmansyah@codeid.co.id", "gerry.wijaya@codeid.co.id"}},
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	gormDB, err := db.OpenGORM(cfg)
	if err != nil {
		log.Fatal(err)
	}

	hasher := auth.NewPasswordHasher()

	if err := gormDB.Transaction(func(tx *gorm.DB) error {
		organization, err := ensureOrganization(tx)
		if err != nil {
			return err
		}

		roles, err := ensureRoles(tx)
		if err != nil {
			return err
		}

		jobLevels, err := ensureJobLevels(tx, organization.ID)
		if err != nil {
			return err
		}
		employeeTypes, err := ensureEmployeeTypes(tx, organization.ID)
		if err != nil {
			return err
		}
		leaveTypes, err := ensureLeaveTypes(tx, organization.ID)
		if err != nil {
			return err
		}
		employmentStatuses, err := ensureEmploymentStatuses(tx, organization.ID)
		if err != nil {
			return err
		}
		locations, err := ensureLocations(tx, organization.ID)
		if err != nil {
			return err
		}
		departments, err := ensureDepartments(tx, organization.ID, locations)
		if err != nil {
			return err
		}
		jobs, err := ensureJobs(tx, organization.ID, departments, jobLevels)
		if err != nil {
			return err
		}
		if err := ensureHolidays(tx, organization.ID, locations); err != nil {
			return err
		}
		if err := ensureAdminUser(tx, cfg, hasher, organization.ID, roles["admin"], employeeTypes, employmentStatuses, departments, jobs, locations); err != nil {
			return err
		}
		employees, err := ensureEmployees(tx, hasher, organization.ID, roles, employeeTypes, employmentStatuses, departments, jobs, locations)
		if err != nil {
			return err
		}
		if err := attachDepartmentManagers(tx, departments, employees); err != nil {
			return err
		}
		if err := ensureTeams(tx, organization.ID, departments, employees); err != nil {
			return err
		}
		if err := ensureRoleAssignments(tx, organization.ID, employees, departments, jobs); err != nil {
			return err
		}
		if err := ensureAttendance(tx, organization.ID, employees); err != nil {
			return err
		}
		if err := ensureLeaveRequests(tx, organization.ID, employees, leaveTypes); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	log.Println("seeded PT. CODEID reference and demo data")
}

func ensureOrganization(tx *gorm.DB) (models.Organization, error) {
	var organization models.Organization
	err := tx.Where("code = ?", shared.DefaultOrganizationCode).
		Attrs(models.Organization{ID: uuid.New(), Code: shared.DefaultOrganizationCode, Name: "PT. CODEID"}).
		FirstOrCreate(&organization).Error
	return organization, err
}

func ensureRoles(tx *gorm.DB) (map[string]models.Role, error) {
	codes := []string{"admin", "manager", "employee"}
	roles := map[string]models.Role{}
	for _, code := range codes {
		var role models.Role
		if err := tx.Where("code = ?", code).First(&role).Error; err != nil {
			return nil, err
		}
		roles[code] = role
	}
	return roles, nil
}

func ensureJobLevels(tx *gorm.DB, organizationID uuid.UUID) (map[string]models.JobLevel, error) {
	items := []models.JobLevel{
		{OrganizationID: organizationID, Code: "entry", Name: "Entry", Rank: 1},
		{OrganizationID: organizationID, Code: "junior", Name: "Junior", Rank: 2},
		{OrganizationID: organizationID, Code: "mid", Name: "Mid", Rank: 3},
		{OrganizationID: organizationID, Code: "senior", Name: "Senior", Rank: 4},
		{OrganizationID: organizationID, Code: "lead", Name: "Lead", Rank: 5},
		{OrganizationID: organizationID, Code: "manager", Name: "Manager", Rank: 6},
	}
	result := map[string]models.JobLevel{}
	for _, item := range items {
		var current models.JobLevel
		if err := tx.Where("organization_id = ? AND code = ?", organizationID, item.Code).
			Assign(map[string]any{"name": item.Name, "rank": item.Rank}).
			Attrs(models.JobLevel{ID: uuid.New(), OrganizationID: organizationID, Code: item.Code, Name: item.Name, Rank: item.Rank}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[current.Code] = current
	}
	return result, nil
}

func ensureEmployeeTypes(tx *gorm.DB, organizationID uuid.UUID) (map[string]models.EmployeeType, error) {
	items := []models.EmployeeType{
		{OrganizationID: organizationID, Code: "full-time", Name: "Full-time"},
		{OrganizationID: organizationID, Code: "part-time", Name: "Part-time"},
		{OrganizationID: organizationID, Code: "contract", Name: "Contract"},
		{OrganizationID: organizationID, Code: "freelance", Name: "Freelance"},
		{OrganizationID: organizationID, Code: "internship", Name: "Internship"},
	}
	result := map[string]models.EmployeeType{}
	for _, item := range items {
		var current models.EmployeeType
		if err := tx.Where("organization_id = ? AND code = ?", organizationID, item.Code).
			Assign(map[string]any{"name": item.Name}).
			Attrs(models.EmployeeType{ID: uuid.New(), OrganizationID: organizationID, Code: item.Code, Name: item.Name}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[current.Code] = current
	}
	return result, nil
}

func ensureLeaveTypes(tx *gorm.DB, organizationID uuid.UUID) (map[string]models.LeaveType, error) {
	items := []models.LeaveType{
		{OrganizationID: organizationID, Code: "annual", Name: "Annual Leave", Description: "Planned annual leave entitlement", IsPaid: true},
		{OrganizationID: organizationID, Code: "sick", Name: "Sick Leave", Description: "Medical or health-related absence", IsPaid: true},
		{OrganizationID: organizationID, Code: "emergency", Name: "Emergency Leave", Description: "Urgent unforeseen personal matters", IsPaid: true},
		{OrganizationID: organizationID, Code: "maternity-paternity", Name: "Maternity/Paternity Leave", Description: "Family care and birth leave", IsPaid: true},
		{OrganizationID: organizationID, Code: "unpaid", Name: "Unpaid Leave", Description: "Approved unpaid leave", IsPaid: false},
	}
	result := map[string]models.LeaveType{}
	for _, item := range items {
		var current models.LeaveType
		if err := tx.Where("organization_id = ? AND code = ?", organizationID, item.Code).
			Assign(map[string]any{"name": item.Name, "description": item.Description, "is_paid": item.IsPaid}).
			Attrs(models.LeaveType{ID: uuid.New(), OrganizationID: organizationID, Code: item.Code, Name: item.Name, Description: item.Description, IsPaid: item.IsPaid}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[current.Code] = current
	}
	return result, nil
}

func ensureEmploymentStatuses(tx *gorm.DB, organizationID uuid.UUID) (map[string]models.EmploymentStatus, error) {
	items := []models.EmploymentStatus{
		{OrganizationID: organizationID, Code: "active", Name: "Active"},
		{OrganizationID: organizationID, Code: "probation", Name: "Probation"},
		{OrganizationID: organizationID, Code: "on-leave", Name: "On Leave"},
		{OrganizationID: organizationID, Code: "resigned", Name: "Resigned"},
		{OrganizationID: organizationID, Code: "terminated", Name: "Terminated"},
	}
	result := map[string]models.EmploymentStatus{}
	for _, item := range items {
		var current models.EmploymentStatus
		if err := tx.Where("organization_id = ? AND code = ?", organizationID, item.Code).
			Assign(map[string]any{"name": item.Name}).
			Attrs(models.EmploymentStatus{ID: uuid.New(), OrganizationID: organizationID, Code: item.Code, Name: item.Name}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[current.Code] = current
	}
	return result, nil
}
func ensureLocations(tx *gorm.DB, organizationID uuid.UUID) (map[string]models.Location, error) {
	items := []models.Location{
		{OrganizationID: organizationID, Name: "Jakarta Headquarters", Address: "Jl. Jenderal Sudirman No. 10", City: "Jakarta", Notes: "PT. CODEID main office"},
		{OrganizationID: organizationID, Name: "Bandung Branch", Address: "Jl. Dago No. 25", City: "Bandung", Notes: "Bootcamp branch office"},
		{OrganizationID: organizationID, Name: "Remote / Client Location", Address: "Client Locations", City: "Remote", Notes: "For remote or onsite client delivery"},
	}
	result := map[string]models.Location{}
	for _, item := range items {
		var current models.Location
		if err := tx.Where("organization_id = ? AND name = ?", organizationID, item.Name).
			Assign(map[string]any{"address": item.Address, "city": item.City, "notes": item.Notes}).
			Attrs(models.Location{ID: uuid.New(), OrganizationID: organizationID, Name: item.Name, Address: item.Address, City: item.City, Notes: item.Notes}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[item.Name] = current
	}
	return result, nil
}

func ensureDepartments(tx *gorm.DB, organizationID uuid.UUID, locations map[string]models.Location) (map[string]models.Department, error) {
	seeds := []departmentSeed{
		{Name: "PT. CODEID", Location: "Jakarta Headquarters", Level: 0},
		{Name: "Divisi IT", ParentName: "PT. CODEID", Location: "Jakarta Headquarters", Level: 1},
		{Name: "Divisi Business & Development", ParentName: "PT. CODEID", Location: "Jakarta Headquarters", Level: 1},
		{Name: "Divisi Sales", ParentName: "PT. CODEID", Location: "Jakarta Headquarters", Level: 1},
		{Name: "Divisi Finance", ParentName: "PT. CODEID", Location: "Jakarta Headquarters", Level: 1},
		{Name: "Divisi HR", ParentName: "PT. CODEID", Location: "Jakarta Headquarters", Level: 1},
		{Name: "Infrastructure", ParentName: "Divisi IT", Location: "Jakarta Headquarters", Level: 2},
		{Name: "System Development", ParentName: "Divisi IT", Location: "Jakarta Headquarters", Level: 2},
		{Name: "Security", ParentName: "Divisi IT", Location: "Jakarta Headquarters", Level: 2},
		{Name: "Smartsourcing", ParentName: "Divisi Business & Development", Location: "Remote / Client Location", Level: 2},
		{Name: "Bootcamp", ParentName: "Divisi Business & Development", Location: "Bandung Branch", Level: 2},
	}
	result := map[string]models.Department{}
	for _, seed := range seeds {
		department := models.Department{OrganizationID: organizationID, Name: seed.Name, LocationID: locations[seed.Location].ID, Level: seed.Level}
		if seed.ParentName != "" {
			parent := result[seed.ParentName]
			department.ParentDepartmentID = &parent.ID
		}
		var current models.Department
		if err := tx.Where("organization_id = ? AND name = ?", organizationID, department.Name).
			Assign(map[string]any{"location_id": department.LocationID, "level": department.Level, "parent_department_id": department.ParentDepartmentID}).
			Attrs(models.Department{ID: uuid.New(), OrganizationID: organizationID, Name: department.Name, LocationID: department.LocationID, Level: department.Level, ParentDepartmentID: department.ParentDepartmentID}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[seed.Name] = current
	}
	return result, nil
}

func ensureJobs(tx *gorm.DB, organizationID uuid.UUID, departments map[string]models.Department, levels map[string]models.JobLevel) (map[string]models.Job, error) {
	seeds := []jobSeed{
		{Title: "IT Manager", Department: "Divisi IT", Level: "manager", MinSalary: 18000000, MaxSalary: 26000000, Description: "Leads IT strategy, delivery, and infrastructure governance."},
		{Title: "Infrastructure Lead", Department: "Infrastructure", Level: "lead", MinSalary: 15500000, MaxSalary: 21500000, Description: "Coordinates infrastructure reliability, endpoint readiness, and office technology operations."},
		{Title: "System Development Lead", Department: "System Development", Level: "lead", MinSalary: 16500000, MaxSalary: 22500000, Description: "Leads the system development subdepartment and coordinates engineering execution."},
		{Title: "Junior System Developer", Department: "System Development", Level: "junior", MinSalary: 8000000, MaxSalary: 12000000, Description: "Supports software delivery while building core engineering capabilities."},
		{Title: "System Developer", Department: "System Development", Level: "mid", MinSalary: 12000000, MaxSalary: 20000000, Description: "Builds and maintains internal and client-facing software systems."},
		{Title: "Senior System Developer", Department: "System Development", Level: "senior", MinSalary: 15000000, MaxSalary: 23000000, Description: "Leads technical implementation and mentors junior engineers on delivery quality."},
		{Title: "Security Specialist", Department: "Security", Level: "senior", MinSalary: 15000000, MaxSalary: 23000000, Description: "Owns security reviews, controls, and incident preparedness."},
		{Title: "Network Engineer", Department: "Infrastructure", Level: "mid", MinSalary: 11000000, MaxSalary: 18000000, Description: "Maintains network reliability and office connectivity."},
		{Title: "Business Development Manager", Department: "Divisi Business & Development", Level: "manager", MinSalary: 19000000, MaxSalary: 28000000, Description: "Leads growth initiatives, partnerships, and new business channels."},
		{Title: "Bootcamp Lead", Department: "Bootcamp", Level: "lead", MinSalary: 15000000, MaxSalary: 21000000, Description: "Leads the bootcamp subdepartment and coordinates day-to-day academy operations."},
		{Title: "Bootcamp Program Coordinator", Department: "Bootcamp", Level: "mid", MinSalary: 8500000, MaxSalary: 13500000, Description: "Coordinates cohort schedules, facilitators, and participant operations."},
		{Title: "Talent Acquisition Specialist", Department: "Bootcamp", Level: "senior", MinSalary: 10000000, MaxSalary: 16000000, Description: "Sources and coordinates hiring demand across client and internal roles."},
		{Title: "Smartsourcing Lead", Department: "Smartsourcing", Level: "lead", MinSalary: 15000000, MaxSalary: 22000000, Description: "Leads client staffing readiness, consultant deployment, and account execution quality."},
		{Title: "Outsourcing Specialist", Department: "Smartsourcing", Level: "mid", MinSalary: 9500000, MaxSalary: 15000000, Description: "Manages outsourcing delivery pipelines and client staffing operations."},
		{Title: "Client Success Analyst", Department: "Smartsourcing", Level: "mid", MinSalary: 9000000, MaxSalary: 14500000, Description: "Supports post-deployment coordination, client health checks, and consultant follow-through."},
		{Title: "Sales Manager", Department: "Divisi Sales", Level: "manager", MinSalary: 17000000, MaxSalary: 25000000, Description: "Leads revenue execution and account growth strategy."},
		{Title: "Sales Team Lead", Department: "Divisi Sales", Level: "lead", MinSalary: 13500000, MaxSalary: 19000000, Description: "Owns daily pipeline review, coaching, and execution support for the sales pod."},
		{Title: "Account Executive", Department: "Divisi Sales", Level: "mid", MinSalary: 9000000, MaxSalary: 15000000, Description: "Owns customer acquisition and account conversion activity."},
		{Title: "Sales Consultant", Department: "Divisi Sales", Level: "junior", MinSalary: 7500000, MaxSalary: 12000000, Description: "Supports lead qualification and consultative product pitching."},
		{Title: "Finance Manager", Department: "Divisi Finance", Level: "manager", MinSalary: 18000000, MaxSalary: 26000000, Description: "Owns planning, accounting controls, and financial reporting."},
		{Title: "Finance Analyst", Department: "Divisi Finance", Level: "junior", MinSalary: 8000000, MaxSalary: 12000000, Description: "Supports reporting packs, reconciliations, and finance operations analysis."},
		{Title: "Accountant", Department: "Divisi Finance", Level: "mid", MinSalary: 9500000, MaxSalary: 14500000, Description: "Maintains ledgers, reconciliations, and monthly close tasks."},
		{Title: "Tax Specialist", Department: "Divisi Finance", Level: "senior", MinSalary: 12000000, MaxSalary: 17000000, Description: "Handles tax compliance and regulatory submissions."},
		{Title: "HR Manager", Department: "Divisi HR", Level: "manager", MinSalary: 16000000, MaxSalary: 24000000, Description: "Leads people operations, policies, and workforce planning."},
		{Title: "People Operations Specialist", Department: "Divisi HR", Level: "mid", MinSalary: 9500000, MaxSalary: 14500000, Description: "Supports employee lifecycle administration, culture initiatives, and internal HR operations."},
		{Title: "Recruitment Specialist", Department: "Divisi HR", Level: "mid", MinSalary: 9000000, MaxSalary: 14000000, Description: "Coordinates hiring workflows and candidate experience."},
		{Title: "Payroll Specialist", Department: "Divisi HR", Level: "mid", MinSalary: 9500000, MaxSalary: 15000000, Description: "Administers payroll preparation and employee compensation records."},
	}
	result := map[string]models.Job{}
	for _, seed := range seeds {
		job := models.Job{
			OrganizationID:      organizationID,
			Title:               seed.Title,
			PrimaryDepartmentID: departments[seed.Department].ID,
			JobLevelID:          levels[seed.Level].ID,
			MinSalary:           floatPointer(seed.MinSalary),
			MaxSalary:           floatPointer(seed.MaxSalary),
			JobDescription:      seed.Description,
		}
		var current models.Job
		if err := tx.Where("organization_id = ? AND title = ?", organizationID, job.Title).
			Assign(map[string]any{"primary_department_id": job.PrimaryDepartmentID, "job_level_id": job.JobLevelID, "min_salary": job.MinSalary, "max_salary": job.MaxSalary, "job_description": job.JobDescription}).
			Attrs(models.Job{ID: uuid.New(), OrganizationID: organizationID, Title: job.Title, PrimaryDepartmentID: job.PrimaryDepartmentID, JobLevelID: job.JobLevelID, MinSalary: job.MinSalary, MaxSalary: job.MaxSalary, JobDescription: job.JobDescription}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}
		result[job.Title] = current
	}
	return result, nil
}

func ensureHolidays(tx *gorm.DB, organizationID uuid.UUID, locations map[string]models.Location) error {
	year := time.Now().UTC().Year()
	items := []struct {
		Date     string
		Name     string
		Location string
		Notes    string
	}{
		{Date: fmtDate(year, 1, 1), Name: "New Year's Day", Notes: "National holiday observed across PT. CODEID locations."},
		{Date: fmtDate(year, 1, 29), Name: "Chinese New Year", Notes: "National holiday for all operating units."},
		{Date: fmtDate(year, 3, 29), Name: "Nyepi", Notes: "National day of silence with no operational attendance expected."},
		{Date: fmtDate(year, 3, 31), Name: "Company Wellness Reset", Location: "Remote / Client Location", Notes: "Custom company recharge day for remote and client-based contributors."},
		{Date: fmtDate(year, 4, 1), Name: "Eid al-Fitr Holiday", Notes: "National holiday period."},
		{Date: fmtDate(year, 4, 18), Name: "Bootcamp Showcase Day", Location: "Bandung Branch", Notes: "Bandung branch operational closure for cohort showcase and partner event."},
		{Date: fmtDate(year, 5, 1), Name: "Labour Day", Notes: "National labour day."},
		{Date: fmtDate(year, 6, 1), Name: "Pancasila Day", Notes: "National holiday observed organization-wide."},
		{Date: fmtDate(year, 7, 11), Name: "Jakarta HQ Maintenance Break", Location: "Jakarta Headquarters", Notes: "Jakarta-only office closure for planned infrastructure maintenance."},
		{Date: fmtDate(year, 8, 17), Name: "Indonesia Independence Day", Notes: "National public holiday."},
		{Date: fmtDate(year, 9, 26), Name: "PT. CODEID Anniversary", Notes: "Custom company holiday celebrating the organization anniversary."},
		{Date: fmtDate(year, 12, 25), Name: "Christmas Day", Notes: "Observed holiday."},
		{Date: fmtDate(year, 12, 31), Name: "Year-End Company Shutdown", Notes: "Custom company shutdown day before year close handover."},
	}
	for _, seed := range items {
		holidayDate, err := time.Parse("2006-01-02", seed.Date)
		if err != nil {
			return err
		}
		var locationID *uuid.UUID
		if seed.Location != "" {
			value := locations[seed.Location].ID
			locationID = &value
		}
		holiday := models.Holiday{OrganizationID: organizationID, HolidayDate: holidayDate, Name: seed.Name, LocationID: locationID, Notes: seed.Notes}
		var current models.Holiday
		if err := tx.Where("organization_id = ? AND holiday_date = ? AND name = ?", organizationID, holidayDate, seed.Name).
			Assign(map[string]any{"location_id": holiday.LocationID, "notes": holiday.Notes}).
			Attrs(models.Holiday{ID: uuid.New(), OrganizationID: organizationID, HolidayDate: holidayDate, Name: seed.Name, LocationID: holiday.LocationID, Notes: holiday.Notes}).
			FirstOrCreate(&current).Error; err != nil {
			return err
		}
	}
	return nil
}
func ensureAdminUser(
	tx *gorm.DB,
	cfg config.Config,
	hasher auth.PasswordHasher,
	organizationID uuid.UUID,
	role models.Role,
	employeeTypes map[string]models.EmployeeType,
	employmentStatuses map[string]models.EmploymentStatus,
	departments map[string]models.Department,
	jobs map[string]models.Job,
	locations map[string]models.Location,
) error {
	email := strings.ToLower(cfg.AdminSeedEmail)
	if email == "" {
		email = "admin@codeid.local"
	}
	password := cfg.AdminSeedPassword
	if password == "" {
		password = "ChangeMe123!"
	}
	user, err := ensureUser(tx, hasher, organizationID, role, email, password, nil)
	if err != nil {
		return err
	}

	hireDate := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	var employee models.Employee
	if err := tx.Where("user_id = ? OR employee_code = ?", user.ID, "CID-ADMIN-001").
		Assign(map[string]any{
			"organization_id":        organizationID,
			"user_id":                user.ID,
			"first_name":             "System",
			"last_name":              "Administrator",
			"email":                  email,
			"phone_number":           "+62-21-0000-0000",
			"hire_date":              hireDate,
			"employee_type_id":       employeeTypes["full-time"].ID,
			"employment_status_id":   employmentStatuses["active"].ID,
			"department_id":          departments["Divisi HR"].ID,
			"job_id":                 jobs["HR Manager"].ID,
			"location_id":            locations["Jakarta Headquarters"].ID,
			"work_mode":              "onsite",
			"management_scope":       shared.ManagementScopeIndividual,
			"manager_employee_id":    nil,
		}).
		Attrs(models.Employee{
			ID:                 uuid.New(),
			OrganizationID:     organizationID,
			EmployeeCode:       "CID-ADMIN-001",
			UserID:             &user.ID,
			FirstName:          "System",
			LastName:           "Administrator",
			Email:              email,
			PhoneNumber:        "+62-21-0000-0000",
			HireDate:           hireDate,
			EmployeeTypeID:     employeeTypes["full-time"].ID,
			EmploymentStatusID: employmentStatuses["active"].ID,
			DepartmentID:       departments["Divisi HR"].ID,
			JobID:              jobs["HR Manager"].ID,
			LocationID:         locations["Jakarta Headquarters"].ID,
			WorkMode:           "onsite",
			ManagementScope:    shared.ManagementScopeIndividual,
		}).
		FirstOrCreate(&employee).Error; err != nil {
		return err
	}

	return tx.Model(&models.User{}).Where("id = ?", user.ID).Update("employee_id", employee.ID).Error
}

func ensureEmployees(tx *gorm.DB, hasher auth.PasswordHasher, organizationID uuid.UUID, roles map[string]models.Role, employeeTypes map[string]models.EmployeeType, employmentStatuses map[string]models.EmploymentStatus, departments map[string]models.Department, jobs map[string]models.Job, locations map[string]models.Location) (map[string]models.Employee, error) {
	result := map[string]models.Employee{}
	for _, seed := range employeeSeeds {
		role := roles[seed.Role]
		user, err := ensureUser(tx, hasher, organizationID, role, seed.Email, seed.Password, nil)
		if err != nil {
			return nil, err
		}

		hireDate, err := time.Parse("2006-01-02", seed.HireDate)
		if err != nil {
			return nil, err
		}
		employee := models.Employee{
			OrganizationID:     organizationID,
			EmployeeCode:       seed.EmployeeCode,
			UserID:             &user.ID,
			FirstName:          seed.FirstName,
			LastName:           seed.LastName,
			Email:              strings.ToLower(seed.Email),
			PhoneNumber:        seed.Phone,
			HireDate:           hireDate,
			Salary:             floatPointer(seed.Salary),
			EmployeeTypeID:     employeeTypes[seed.EmployeeType].ID,
			EmploymentStatusID: employmentStatuses[seed.EmploymentStatus].ID,
			DepartmentID:       departments[seed.Department].ID,
			JobID:              jobs[seed.Job].ID,
			LocationID:         locations[seed.Location].ID,
			WorkMode:           seed.WorkMode,
			ManagementScope:    seed.ManagementScope,
		}
		if seed.ManagerEmail != "" {
			manager, ok := result[strings.ToLower(seed.ManagerEmail)]
			if !ok {
				return nil, errors.New("manager record must be seeded before direct reports")
			}
			employee.ManagerEmployeeID = &manager.ID
		}

		var current models.Employee
		if err := tx.Where("employee_code = ?", employee.EmployeeCode).
			Assign(map[string]any{"organization_id": organizationID, "user_id": employee.UserID, "first_name": employee.FirstName, "last_name": employee.LastName, "email": employee.Email, "phone_number": employee.PhoneNumber, "hire_date": employee.HireDate, "salary": employee.Salary, "employee_type_id": employee.EmployeeTypeID, "employment_status_id": employee.EmploymentStatusID, "department_id": employee.DepartmentID, "job_id": employee.JobID, "location_id": employee.LocationID, "work_mode": employee.WorkMode, "management_scope": employee.ManagementScope, "manager_employee_id": employee.ManagerEmployeeID}).
			Attrs(models.Employee{ID: uuid.New(), OrganizationID: organizationID, EmployeeCode: employee.EmployeeCode, UserID: employee.UserID, FirstName: employee.FirstName, LastName: employee.LastName, Email: employee.Email, PhoneNumber: employee.PhoneNumber, HireDate: employee.HireDate, Salary: employee.Salary, EmployeeTypeID: employee.EmployeeTypeID, EmploymentStatusID: employee.EmploymentStatusID, DepartmentID: employee.DepartmentID, JobID: employee.JobID, LocationID: employee.LocationID, WorkMode: employee.WorkMode, ManagementScope: employee.ManagementScope, ManagerEmployeeID: employee.ManagerEmployeeID}).
			FirstOrCreate(&current).Error; err != nil {
			return nil, err
		}

		if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("employee_id", current.ID).Error; err != nil {
			return nil, err
		}

		result[strings.ToLower(seed.Email)] = current
	}
	return result, nil
}

func attachDepartmentManagers(tx *gorm.DB, departments map[string]models.Department, employees map[string]models.Employee) error {
	assignments := map[string]string{
		"Divisi IT":                     "aria.pratama@codeid.co.id",
		"Infrastructure":                "hanif.ardana@codeid.co.id",
		"System Development":            "salma.nuraini@codeid.co.id",
		"Security":                      "aria.pratama@codeid.co.id",
		"Divisi Business & Development": "sinta.lestari@codeid.co.id",
		"Smartsourcing":                 "nadia.hapsari@codeid.co.id",
		"Bootcamp":                      "luthfi.ramadhan@codeid.co.id",
		"Divisi Sales":                  "bayu.saputra@codeid.co.id",
		"Divisi Finance":                "rina.wulandari@codeid.co.id",
		"Divisi HR":                     "dimas.kusuma@codeid.co.id",
	}
	for departmentName, managerEmail := range assignments {
		department := departments[departmentName]
		manager := employees[managerEmail]
		if err := tx.Model(&models.Department{}).Where("id = ?", department.ID).Update("manager_employee_id", manager.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureRoleAssignments(tx *gorm.DB, organizationID uuid.UUID, employees map[string]models.Employee, departments map[string]models.Department, jobs map[string]models.Job) error {
	for _, seed := range employeeSeeds {
		employee, ok := employees[strings.ToLower(seed.Email)]
		if !ok {
			continue
		}
		for _, secondary := range seed.Secondary {
			startDate, err := time.Parse("2006-01-02", secondary.StartDate)
			if err != nil {
				return err
			}
			var endDate *time.Time
			if secondary.EndDate != "" {
				value, err := time.Parse("2006-01-02", secondary.EndDate)
				if err != nil {
					return err
				}
				endDate = &value
			}
			assignment := models.EmployeeRoleAssignment{OrganizationID: organizationID, EmployeeID: employee.ID, JobID: jobs[secondary.Job].ID, DepartmentID: departments[secondary.Department].ID, EstimatedHoursPerWeek: secondary.HoursPerWeek, StartDate: startDate, EndDate: endDate, Notes: secondary.Notes}
			var current models.EmployeeRoleAssignment
			if err := tx.Where("employee_id = ? AND job_id = ? AND department_id = ? AND start_date = ?", assignment.EmployeeID, assignment.JobID, assignment.DepartmentID, assignment.StartDate).
				Assign(map[string]any{"organization_id": organizationID, "estimated_hours_per_week": assignment.EstimatedHoursPerWeek, "end_date": assignment.EndDate, "notes": assignment.Notes}).
				Attrs(models.EmployeeRoleAssignment{ID: uuid.New(), OrganizationID: organizationID, EmployeeID: assignment.EmployeeID, JobID: assignment.JobID, DepartmentID: assignment.DepartmentID, EstimatedHoursPerWeek: assignment.EstimatedHoursPerWeek, StartDate: assignment.StartDate, EndDate: assignment.EndDate, Notes: assignment.Notes}).
				FirstOrCreate(&current).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureTeams(tx *gorm.DB, organizationID uuid.UUID, departments map[string]models.Department, employees map[string]models.Employee) error {
	for _, seed := range teamSeeds {
		manager, ok := employees[strings.ToLower(seed.ManagerEmail)]
		if !ok {
			return errors.New("team manager must exist before team seeding")
		}
		department := departments[seed.Department]

		var team models.Team
		if err := tx.Where("organization_id = ? AND name = ?", organizationID, seed.Name).
			Assign(map[string]any{
				"department_id":         department.ID,
				"manager_employee_id":  manager.ID,
				"is_cross_department":  seed.IsCrossDepartment,
				"focus_area":           seed.FocusArea,
			}).
			Attrs(models.Team{
				ID:                uuid.New(),
				OrganizationID:    organizationID,
				Name:              seed.Name,
				DepartmentID:      &department.ID,
				ManagerEmployeeID: manager.ID,
				IsCrossDepartment: seed.IsCrossDepartment,
				FocusArea:         seed.FocusArea,
			}).
			FirstOrCreate(&team).Error; err != nil {
			return err
		}

		for _, memberEmail := range seed.Members {
			member, ok := employees[strings.ToLower(memberEmail)]
			if !ok {
				return errors.New("team member must exist before membership seeding")
			}
			startDate := time.Now().UTC().Truncate(24 * time.Hour)
			var membership models.TeamMembership
			if err := tx.Where("team_id = ? AND employee_id = ?", team.ID, member.ID).
				Assign(map[string]any{
					"organization_id": organizationID,
					"role_name":       "Member",
					"start_date":      startDate,
					"end_date":        nil,
					"notes":           "Seeded operational team membership.",
				}).
				Attrs(models.TeamMembership{
					ID:             uuid.New(),
					OrganizationID: organizationID,
					TeamID:         team.ID,
					EmployeeID:     member.ID,
					RoleName:       "Member",
					StartDate:      startDate,
					Notes:          "Seeded operational team membership.",
				}).
				FirstOrCreate(&membership).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureAttendance(tx *gorm.DB, organizationID uuid.UUID, employees map[string]models.Employee) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	entries := []struct {
		Email    string
		CheckIn  string
		CheckOut string
		Status   string
		Notes    string
	}{
		{Email: "aria.pratama@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:01:00Z", CheckOut: today.Format("2006-01-02") + "T17:12:00Z", Status: "on time", Notes: "Reviewed infrastructure sprint board."},
		{Email: "sinta.lestari@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:05:00Z", CheckOut: today.Format("2006-01-02") + "T17:20:00Z", Status: "on time", Notes: "Client growth review day."},
		{Email: "hanif.ardana@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T07:59:00Z", CheckOut: today.Format("2006-01-02") + "T17:08:00Z", Status: "on time", Notes: "Infrastructure uptime review and vendor coordination."},
		{Email: "nadia.hapsari@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:04:00Z", CheckOut: today.Format("2006-01-02") + "T17:22:00Z", Status: "remote", Notes: "Client staffing review across active accounts."},
		{Email: "fajar.maulana@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:03:00Z", CheckOut: today.Format("2006-01-02") + "T17:16:00Z", Status: "on time", Notes: "Operational team standup and release coordination."},
		{Email: "mira.anggraeni@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:07:00Z", CheckOut: today.Format("2006-01-02") + "T17:11:00Z", Status: "on time", Notes: "Bootcamp operations planning day."},
		{Email: "dewi.pangestu@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:08:00Z", CheckOut: today.Format("2006-01-02") + "T17:13:00Z", Status: "on time", Notes: "Revenue pod coaching and outbound planning."},
		{Email: "nabila.putri@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:10:00Z", CheckOut: today.Format("2006-01-02") + "T17:08:00Z", Status: "remote", Notes: "Working on shared delivery milestone."},
		{Email: "raka.pradipta@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:14:00Z", CheckOut: today.Format("2006-01-02") + "T17:04:00Z", Status: "on time", Notes: "Backend integration support."},
		{Email: "bima.satrya@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:02:00Z", CheckOut: today.Format("2006-01-02") + "T17:26:00Z", Status: "remote", Notes: "Architecture review plus mentoring support."},
		{Email: "intan.maharani@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:20:00Z", CheckOut: today.Format("2006-01-02") + "T17:00:00Z", Status: "late", Notes: "Intern workshop morning before joining sprint tasks."},
		{Email: "dian.permata@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T07:56:00Z", CheckOut: today.Format("2006-01-02") + "T17:02:00Z", Status: "on time", Notes: "Office network maintenance."},
		{Email: "claudia.suryani@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:06:00Z", CheckOut: today.Format("2006-01-02") + "T17:07:00Z", Status: "on time", Notes: "Workspace equipment and connectivity checks."},
		{Email: "farhan.akbar@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:18:00Z", Status: "late", Notes: "Security review follow-up pending."},
		{Email: "ryan.hadikusuma@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:09:00Z", CheckOut: today.Format("2006-01-02") + "T17:04:00Z", Status: "on time", Notes: "Quarterly vulnerability triage support."},
		{Email: "keisha.anindya@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:11:00Z", CheckOut: today.Format("2006-01-02") + "T17:03:00Z", Status: "on time", Notes: "Bootcamp talent screening interviews."},
		{Email: "ajeng.kusnadi@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:13:00Z", CheckOut: today.Format("2006-01-02") + "T15:02:00Z", Status: "on time", Notes: "Part-time cohort administration and logistics."},
		{Email: "yoga.prasetyo@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:09:00Z", CheckOut: today.Format("2006-01-02") + "T17:05:00Z", Status: "on time", Notes: "Recruitment sync with hiring teams."},
		{Email: "vina.mahendra@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:12:00Z", CheckOut: today.Format("2006-01-02") + "T17:18:00Z", Status: "remote", Notes: "Payroll validation from hybrid workspace."},
		{Email: "bagas.nugroho@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:05:00Z", CheckOut: today.Format("2006-01-02") + "T17:02:00Z", Status: "on time", Notes: "Budget variance analysis support."},
		{Email: "talia.rahma@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T07:58:00Z", CheckOut: today.Format("2006-01-02") + "T16:55:00Z", Status: "on time", Notes: "Client assignment support."},
		{Email: "aulia.kirana@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:00:00Z", CheckOut: today.Format("2006-01-02") + "T16:50:00Z", Status: "remote", Notes: "Client-site coordination and timesheet sync."},
		{Email: "rio.firmansyah@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:01:00Z", CheckOut: today.Format("2006-01-02") + "T17:09:00Z", Status: "remote", Notes: "Client staffing handoff for strategic account."},
		{Email: "laras.ayu@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:04:00Z", CheckOut: today.Format("2006-01-02") + "T16:47:00Z", Status: "remote", Notes: "Remote client success follow-ups and consultant sync."},
		{Email: "gerry.wijaya@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:16:00Z", CheckOut: today.Format("2006-01-02") + "T17:01:00Z", Status: "late", Notes: "Follow-up on solution demo with client."},
		{Email: "selvi.anggraini@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:07:00Z", CheckOut: today.Format("2006-01-02") + "T17:10:00Z", Status: "on time", Notes: "Sales discovery calls and handoff preparation."},
		{Email: "nico.febriansyah@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:24:00Z", CheckOut: today.Format("2006-01-02") + "T17:00:00Z", Status: "late", Notes: "Probation coaching follow-up on morning pipeline sync."},
		{Email: "putri.ardelia@codeid.co.id", CheckIn: today.Format("2006-01-02") + "T08:04:00Z", CheckOut: today.Format("2006-01-02") + "T17:06:00Z", Status: "on time", Notes: "Month-end accounting close support."},
	}
	for _, entry := range entries {
		employee := employees[entry.Email]
		attendance := models.Attendance{ID: uuid.New(), OrganizationID: organizationID, EmployeeID: employee.ID, AttendanceDate: today, Status: entry.Status, Notes: entry.Notes}
		if entry.CheckIn != "" {
			value, err := time.Parse(time.RFC3339, entry.CheckIn)
			if err != nil {
				return err
			}
			attendance.CheckInAt = &value
		}
		if entry.CheckOut != "" {
			value, err := time.Parse(time.RFC3339, entry.CheckOut)
			if err != nil {
				return err
			}
			attendance.CheckOutAt = &value
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "employee_id"}, {Name: "attendance_date"}}, DoUpdates: clause.AssignmentColumns([]string{"organization_id", "check_in_at", "check_out_at", "status", "notes", "updated_at"})}).Create(&attendance).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureLeaveRequests(tx *gorm.DB, organizationID uuid.UUID, employees map[string]models.Employee, leaveTypes map[string]models.LeaveType) error {
	entries := []struct {
		Email      string
		Approver   string
		LeaveType  string
		StartDate  string
		EndDate    string
		Reason     string
		Status     string
		ApprovedAt string
	}{
		{Email: "nabila.putri@codeid.co.id", Approver: "aria.pratama@codeid.co.id", LeaveType: "annual", StartDate: "2026-04-18", EndDate: "2026-04-19", Reason: "Family travel.", Status: "pending"},
		{Email: "talia.rahma@codeid.co.id", Approver: "sinta.lestari@codeid.co.id", LeaveType: "sick", StartDate: "2026-03-22", EndDate: "2026-03-22", Reason: "Medical recovery day.", Status: "approved", ApprovedAt: "2026-03-21T10:00:00Z"},
		{Email: "rafi.hakim@codeid.co.id", Approver: "bayu.saputra@codeid.co.id", LeaveType: "emergency", StartDate: "2026-04-08", EndDate: "2026-04-08", Reason: "Urgent family matter.", Status: "review"},
		{Email: "fajar.maulana@codeid.co.id", Approver: "salma.nuraini@codeid.co.id", LeaveType: "annual", StartDate: "2026-04-24", EndDate: "2026-04-25", Reason: "Family event outside town.", Status: "pending"},
		{Email: "mira.anggraeni@codeid.co.id", Approver: "luthfi.ramadhan@codeid.co.id", LeaveType: "annual", StartDate: "2026-04-28", EndDate: "2026-04-29", Reason: "Personal leave after bootcamp cycle.", Status: "pending"},
		{Email: "salma.nuraini@codeid.co.id", Approver: "aria.pratama@codeid.co.id", LeaveType: "emergency", StartDate: "2026-03-30", EndDate: "2026-03-30", Reason: "Family medical assistance.", Status: "approved", ApprovedAt: "2026-03-29T09:20:00Z"},
		{Email: "keisha.anindya@codeid.co.id", Approver: "luthfi.ramadhan@codeid.co.id", LeaveType: "annual", StartDate: "2026-04-16", EndDate: "2026-04-17", Reason: "Short family trip.", Status: "rejected"},
		{Email: "yoga.prasetyo@codeid.co.id", Approver: "dimas.kusuma@codeid.co.id", LeaveType: "sick", StartDate: "2026-04-10", EndDate: "2026-04-10", Reason: "Medical rest day.", Status: "approved", ApprovedAt: "2026-04-09T08:15:00Z"},
		{Email: "putri.ardelia@codeid.co.id", Approver: "rina.wulandari@codeid.co.id", LeaveType: "annual", StartDate: "2026-05-06", EndDate: "2026-05-07", Reason: "Family ceremony.", Status: "pending"},
		{Email: "aulia.kirana@codeid.co.id", Approver: "sinta.lestari@codeid.co.id", LeaveType: "unpaid", StartDate: "2026-04-20", EndDate: "2026-04-21", Reason: "Client relocation support window.", Status: "cancelled"},
		{Email: "bima.satrya@codeid.co.id", Approver: "salma.nuraini@codeid.co.id", LeaveType: "annual", StartDate: "2026-05-12", EndDate: "2026-05-20", Reason: "Requested extended annual leave after already consuming most team quota.", Status: "rejected"},
		{Email: "intan.maharani@codeid.co.id", Approver: "fajar.maulana@codeid.co.id", LeaveType: "sick", StartDate: "2026-04-14", EndDate: "2026-04-14", Reason: "Recovery after internship campus visit.", Status: "pending"},
		{Email: "claudia.suryani@codeid.co.id", Approver: "hanif.ardana@codeid.co.id", LeaveType: "emergency", StartDate: "2026-04-11", EndDate: "2026-04-11", Reason: "Urgent parent care support.", Status: "approved", ApprovedAt: "2026-04-10T09:45:00Z"},
		{Email: "rio.firmansyah@codeid.co.id", Approver: "nadia.hapsari@codeid.co.id", LeaveType: "unpaid", StartDate: "2026-05-03", EndDate: "2026-05-08", Reason: "Client relocation overlap requiring personal travel arrangement.", Status: "review"},
		{Email: "laras.ayu@codeid.co.id", Approver: "nadia.hapsari@codeid.co.id", LeaveType: "annual", StartDate: "2026-06-02", EndDate: "2026-06-03", Reason: "Personal leave after project wrap-up.", Status: "pending"},
		{Email: "ajeng.kusnadi@codeid.co.id", Approver: "mira.anggraeni@codeid.co.id", LeaveType: "emergency", StartDate: "2026-04-15", EndDate: "2026-04-15", Reason: "Family administrative appointment.", Status: "approved", ApprovedAt: "2026-04-14T07:55:00Z"},
		{Email: "mei.lestari@codeid.co.id", Approver: "dimas.kusuma@codeid.co.id", LeaveType: "maternity-paternity", StartDate: "2026-04-01", EndDate: "2026-04-30", Reason: "Extended maternity recovery period.", Status: "approved", ApprovedAt: "2026-03-28T11:30:00Z"},
		{Email: "bagas.nugroho@codeid.co.id", Approver: "rina.wulandari@codeid.co.id", LeaveType: "annual", StartDate: "2026-04-23", EndDate: "2026-04-24", Reason: "Family event and travel support.", Status: "pending"},
		{Email: "selvi.anggraini@codeid.co.id", Approver: "dewi.pangestu@codeid.co.id", LeaveType: "sick", StartDate: "2026-04-09", EndDate: "2026-04-09", Reason: "Medical check-up recovery.", Status: "approved", ApprovedAt: "2026-04-08T18:05:00Z"},
		{Email: "nico.febriansyah@codeid.co.id", Approver: "dewi.pangestu@codeid.co.id", LeaveType: "annual", StartDate: "2026-05-15", EndDate: "2026-05-16", Reason: "Requested annual leave during probation review window.", Status: "rejected"},
	}
	for _, entry := range entries {
		startDate, err := time.Parse("2006-01-02", entry.StartDate)
		if err != nil {
			return err
		}
		endDate, err := time.Parse("2006-01-02", entry.EndDate)
		if err != nil {
			return err
		}
		leave := models.LeaveRequest{OrganizationID: organizationID, EmployeeID: employees[entry.Email].ID, LeaveTypeID: leaveTypes[entry.LeaveType].ID, StartDate: startDate, EndDate: endDate, Reason: entry.Reason, Status: entry.Status}
		if entry.Approver != "" {
			approverID := employees[entry.Approver].ID
			leave.ApproverEmployeeID = &approverID
		}
		if entry.ApprovedAt != "" {
			approvedAt, err := time.Parse(time.RFC3339, entry.ApprovedAt)
			if err != nil {
				return err
			}
			leave.ApprovedAt = &approvedAt
		}
		var current models.LeaveRequest
		if err := tx.Where("employee_id = ? AND leave_type_id = ? AND start_date = ?", leave.EmployeeID, leave.LeaveTypeID, leave.StartDate).
			Assign(map[string]any{"organization_id": organizationID, "approver_employee_id": leave.ApproverEmployeeID, "end_date": leave.EndDate, "reason": leave.Reason, "status": leave.Status, "approved_at": leave.ApprovedAt}).
			Attrs(models.LeaveRequest{ID: uuid.New(), OrganizationID: organizationID, EmployeeID: leave.EmployeeID, ApproverEmployeeID: leave.ApproverEmployeeID, LeaveTypeID: leave.LeaveTypeID, StartDate: leave.StartDate, EndDate: leave.EndDate, Reason: leave.Reason, Status: leave.Status, ApprovedAt: leave.ApprovedAt}).
			FirstOrCreate(&current).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureUser(tx *gorm.DB, hasher auth.PasswordHasher, organizationID uuid.UUID, role models.Role, email, password string, employeeID *uuid.UUID) (models.User, error) {
	hash, err := hasher.Hash(password)
	if err != nil {
		return models.User{}, err
	}
	var user models.User
	assignments := map[string]any{"password_hash": hash, "role_id": role.ID, "organization_id": organizationID, "is_active": true, "employee_id": employeeID}
	if err := tx.Where("email = ?", strings.ToLower(email)).
		Assign(assignments).
		Attrs(models.User{ID: uuid.New(), OrganizationID: organizationID, Email: strings.ToLower(email)}).
		FirstOrCreate(&user).Error; err != nil {
		return models.User{}, err
	}
	if err := tx.Model(&user).Updates(assignments).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func floatPointer(value float64) *float64 {
	return &value
}

func fmtDate(year, month, day int) string {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}
