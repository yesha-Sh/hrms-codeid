package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Role struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Code string `gorm:"column:code"`
	Name string `gorm:"column:name"`
}

type Organization struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Code      string    `gorm:"column:code"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type JobLevel struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	Code           string       `gorm:"column:code"`
	Name           string       `gorm:"column:name"`
	Rank           int          `gorm:"column:rank"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
}

type EmployeeType struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	Code           string       `gorm:"column:code"`
	Name           string       `gorm:"column:name"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
}

type LeaveType struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	Code           string       `gorm:"column:code"`
	Name           string       `gorm:"column:name"`
	Description    string       `gorm:"column:description"`
	IsPaid         bool         `gorm:"column:is_paid"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
}

type EmploymentStatus struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	Code           string       `gorm:"column:code"`
	Name           string       `gorm:"column:name"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
}

type Location struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	Name           string       `gorm:"column:name"`
	Address        string       `gorm:"column:address"`
	City           string       `gorm:"column:city"`
	Notes          string       `gorm:"column:notes"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
}

type Department struct {
	ID                 uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID     uuid.UUID    `gorm:"column:organization_id"`
	Name               string       `gorm:"column:name"`
	ParentDepartmentID *uuid.UUID   `gorm:"column:parent_department_id"`
	ManagerEmployeeID  *uuid.UUID   `gorm:"column:manager_employee_id"`
	LocationID         uuid.UUID    `gorm:"column:location_id"`
	Level              int          `gorm:"column:level"`
	CreatedAt          time.Time    `gorm:"column:created_at"`
	UpdatedAt          time.Time    `gorm:"column:updated_at"`
	Organization       Organization `gorm:"foreignKey:OrganizationID"`
	ParentDepartment   *Department  `gorm:"foreignKey:ParentDepartmentID"`
	ManagerEmployee    *Employee    `gorm:"foreignKey:ManagerEmployeeID"`
	Location           Location     `gorm:"foreignKey:LocationID"`
}

type Job struct {
	ID                  uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID      uuid.UUID    `gorm:"column:organization_id"`
	Title               string       `gorm:"column:title"`
	PrimaryDepartmentID uuid.UUID    `gorm:"column:primary_department_id"`
	JobLevelID          uuid.UUID    `gorm:"column:job_level_id"`
	MinSalary           *float64     `gorm:"column:min_salary"`
	MaxSalary           *float64     `gorm:"column:max_salary"`
	JobDescription      string       `gorm:"column:job_description"`
	CreatedAt           time.Time    `gorm:"column:created_at"`
	UpdatedAt           time.Time    `gorm:"column:updated_at"`
	Organization        Organization `gorm:"foreignKey:OrganizationID"`
	PrimaryDepartment   Department   `gorm:"foreignKey:PrimaryDepartmentID"`
	JobLevel            JobLevel     `gorm:"foreignKey:JobLevelID"`
}

type User struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	Email          string       `gorm:"column:email"`
	PasswordHash   string       `gorm:"column:password_hash"`
	RoleID         int64        `gorm:"column:role_id"`
	EmployeeID     *uuid.UUID   `gorm:"column:employee_id"`
	IsActive       bool         `gorm:"column:is_active"`
	LastLoginAt    *time.Time   `gorm:"column:last_login_at"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
	Role           Role         `gorm:"foreignKey:RoleID"`
	Employee       *Employee    `gorm:"foreignKey:EmployeeID"`
}

type Employee struct {
	ID                 uuid.UUID                `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID     uuid.UUID                `gorm:"column:organization_id"`
	EmployeeCode       string                   `gorm:"column:employee_code"`
	UserID             *uuid.UUID               `gorm:"column:user_id"`
	FirstName          string                   `gorm:"column:first_name"`
	LastName           string                   `gorm:"column:last_name"`
	Email              string                   `gorm:"column:email"`
	PhoneNumber        string                   `gorm:"column:phone_number"`
	HireDate           time.Time                `gorm:"column:hire_date"`
	Salary             *float64                 `gorm:"column:salary"`
	EmployeeTypeID     uuid.UUID                `gorm:"column:employee_type_id"`
	EmploymentStatusID uuid.UUID                `gorm:"column:employment_status_id"`
	DepartmentID       uuid.UUID                `gorm:"column:department_id"`
	JobID              uuid.UUID                `gorm:"column:job_id"`
	LocationID         uuid.UUID                `gorm:"column:location_id"`
	WorkMode           string                   `gorm:"column:work_mode"`
	ManagementScope    string                   `gorm:"column:management_scope"`
	ManagerEmployeeID  *uuid.UUID               `gorm:"column:manager_employee_id"`
	CreatedAt          time.Time                `gorm:"column:created_at"`
	UpdatedAt          time.Time                `gorm:"column:updated_at"`
	Organization       Organization             `gorm:"foreignKey:OrganizationID"`
	EmployeeType       EmployeeType             `gorm:"foreignKey:EmployeeTypeID"`
	EmploymentStatus   EmploymentStatus         `gorm:"foreignKey:EmploymentStatusID"`
	Department         Department               `gorm:"foreignKey:DepartmentID"`
	Job                Job                      `gorm:"foreignKey:JobID"`
	Location           Location                 `gorm:"foreignKey:LocationID"`
	Manager            *Employee                `gorm:"foreignKey:ManagerEmployeeID"`
	RoleAssignments    []EmployeeRoleAssignment `gorm:"foreignKey:EmployeeID"`
	ManagedTeams       []Team                   `gorm:"foreignKey:ManagerEmployeeID"`
	TeamMemberships    []TeamMembership         `gorm:"foreignKey:EmployeeID"`
}

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	UserAgent string     `gorm:"column:user_agent"`
	IPAddress string     `gorm:"column:ip_address"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

type EmployeeRoleAssignment struct {
	ID                    uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID        uuid.UUID    `gorm:"column:organization_id"`
	EmployeeID            uuid.UUID    `gorm:"column:employee_id"`
	JobID                 uuid.UUID    `gorm:"column:job_id"`
	DepartmentID          uuid.UUID    `gorm:"column:department_id"`
	EstimatedHoursPerWeek float64      `gorm:"column:estimated_hours_per_week"`
	StartDate             time.Time    `gorm:"column:start_date"`
	EndDate               *time.Time   `gorm:"column:end_date"`
	Notes                 string       `gorm:"column:notes"`
	CreatedAt             time.Time    `gorm:"column:created_at"`
	UpdatedAt             time.Time    `gorm:"column:updated_at"`
	Organization          Organization `gorm:"foreignKey:OrganizationID"`
	Employee              Employee     `gorm:"foreignKey:EmployeeID"`
	Job                   Job          `gorm:"foreignKey:JobID"`
	Department            Department   `gorm:"foreignKey:DepartmentID"`
}

type Attendance struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	EmployeeID     uuid.UUID    `gorm:"column:employee_id"`
	AttendanceDate time.Time    `gorm:"column:attendance_date"`
	CheckInAt      *time.Time   `gorm:"column:check_in_at"`
	CheckOutAt     *time.Time   `gorm:"column:check_out_at"`
	Status         string       `gorm:"column:status"`
	Notes          string       `gorm:"column:notes"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
	Employee       Employee     `gorm:"foreignKey:EmployeeID"`
}

type Team struct {
	ID                uuid.UUID        `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID    uuid.UUID        `gorm:"column:organization_id"`
	Name              string           `gorm:"column:name"`
	DepartmentID      *uuid.UUID       `gorm:"column:department_id"`
	ManagerEmployeeID uuid.UUID        `gorm:"column:manager_employee_id"`
	IsCrossDepartment bool             `gorm:"column:is_cross_department"`
	FocusArea         string           `gorm:"column:focus_area"`
	CreatedAt         time.Time        `gorm:"column:created_at"`
	UpdatedAt         time.Time        `gorm:"column:updated_at"`
	Organization      Organization     `gorm:"foreignKey:OrganizationID"`
	Department        *Department      `gorm:"foreignKey:DepartmentID"`
	ManagerEmployee   Employee         `gorm:"foreignKey:ManagerEmployeeID"`
	Memberships       []TeamMembership `gorm:"foreignKey:TeamID"`
}

type TeamMembership struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	TeamID         uuid.UUID    `gorm:"column:team_id"`
	EmployeeID     uuid.UUID    `gorm:"column:employee_id"`
	RoleName       string       `gorm:"column:role_name"`
	StartDate      time.Time    `gorm:"column:start_date"`
	EndDate        *time.Time   `gorm:"column:end_date"`
	Notes          string       `gorm:"column:notes"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
	Team           Team         `gorm:"foreignKey:TeamID"`
	Employee       Employee     `gorm:"foreignKey:EmployeeID"`
}

type Holiday struct {
	ID             uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID    `gorm:"column:organization_id"`
	HolidayDate    time.Time    `gorm:"column:holiday_date"`
	Name           string       `gorm:"column:name"`
	LocationID     *uuid.UUID   `gorm:"column:location_id"`
	Notes          string       `gorm:"column:notes"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
	Location       *Location    `gorm:"foreignKey:LocationID"`
}

type LeaveRequest struct {
	ID                 uuid.UUID    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID     uuid.UUID    `gorm:"column:organization_id"`
	EmployeeID         uuid.UUID    `gorm:"column:employee_id"`
	ApproverEmployeeID *uuid.UUID   `gorm:"column:approver_employee_id"`
	LeaveTypeID        uuid.UUID    `gorm:"column:leave_type_id"`
	StartDate          time.Time    `gorm:"column:start_date"`
	EndDate            time.Time    `gorm:"column:end_date"`
	Reason             string       `gorm:"column:reason"`
	Status             string       `gorm:"column:status"`
	ApprovedAt         *time.Time   `gorm:"column:approved_at"`
	CreatedAt          time.Time    `gorm:"column:created_at"`
	UpdatedAt          time.Time    `gorm:"column:updated_at"`
	Organization       Organization `gorm:"foreignKey:OrganizationID"`
	Employee           Employee     `gorm:"foreignKey:EmployeeID"`
	Approver           *Employee    `gorm:"foreignKey:ApproverEmployeeID"`
	LeaveType          LeaveType    `gorm:"foreignKey:LeaveTypeID"`
}

type AuditLog struct {
	ID             uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID      `gorm:"column:organization_id"`
	ActorUserID    *uuid.UUID     `gorm:"column:actor_user_id"`
	Action         string         `gorm:"column:action"`
	Entity         string         `gorm:"column:entity"`
	EntityID       *uuid.UUID     `gorm:"column:entity_id"`
	Metadata       datatypes.JSON `gorm:"column:metadata_jsonb"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID"`
	ActorUser      *User          `gorm:"foreignKey:ActorUserID"`
}

func (Role) TableName() string                   { return "roles" }
func (Organization) TableName() string           { return "organizations" }
func (JobLevel) TableName() string               { return "job_levels" }
func (EmployeeType) TableName() string           { return "employee_types" }
func (LeaveType) TableName() string              { return "leave_types" }
func (EmploymentStatus) TableName() string       { return "employment_statuses" }
func (Location) TableName() string               { return "locations" }
func (Department) TableName() string             { return "departments" }
func (Job) TableName() string                    { return "jobs" }
func (User) TableName() string                   { return "users" }
func (Employee) TableName() string               { return "employees" }
func (RefreshToken) TableName() string           { return "refresh_tokens" }
func (EmployeeRoleAssignment) TableName() string { return "employee_role_assignments" }
func (Attendance) TableName() string             { return "attendances" }
func (Team) TableName() string                   { return "teams" }
func (TeamMembership) TableName() string         { return "team_memberships" }
func (Holiday) TableName() string                { return "holidays" }
func (LeaveRequest) TableName() string           { return "leave_requests" }
func (AuditLog) TableName() string               { return "audit_logs" }

func ensureUUID(id *uuid.UUID) {
	if id != nil && *id == uuid.Nil {
		*id = uuid.New()
	}
}

func (m *Organization) BeforeCreate(*gorm.DB) error           { ensureUUID(&m.ID); return nil }
func (m *JobLevel) BeforeCreate(*gorm.DB) error               { ensureUUID(&m.ID); return nil }
func (m *EmployeeType) BeforeCreate(*gorm.DB) error           { ensureUUID(&m.ID); return nil }
func (m *LeaveType) BeforeCreate(*gorm.DB) error              { ensureUUID(&m.ID); return nil }
func (m *EmploymentStatus) BeforeCreate(*gorm.DB) error       { ensureUUID(&m.ID); return nil }
func (m *Location) BeforeCreate(*gorm.DB) error               { ensureUUID(&m.ID); return nil }
func (m *Department) BeforeCreate(*gorm.DB) error             { ensureUUID(&m.ID); return nil }
func (m *Job) BeforeCreate(*gorm.DB) error                    { ensureUUID(&m.ID); return nil }
func (m *User) BeforeCreate(*gorm.DB) error                   { ensureUUID(&m.ID); return nil }
func (m *Employee) BeforeCreate(*gorm.DB) error               { ensureUUID(&m.ID); return nil }
func (m *RefreshToken) BeforeCreate(*gorm.DB) error           { ensureUUID(&m.ID); return nil }
func (m *EmployeeRoleAssignment) BeforeCreate(*gorm.DB) error { ensureUUID(&m.ID); return nil }
func (m *Attendance) BeforeCreate(*gorm.DB) error             { ensureUUID(&m.ID); return nil }
func (m *Team) BeforeCreate(*gorm.DB) error                   { ensureUUID(&m.ID); return nil }
func (m *TeamMembership) BeforeCreate(*gorm.DB) error         { ensureUUID(&m.ID); return nil }
func (m *Holiday) BeforeCreate(*gorm.DB) error                { ensureUUID(&m.ID); return nil }
func (m *LeaveRequest) BeforeCreate(*gorm.DB) error           { ensureUUID(&m.ID); return nil }
func (m *AuditLog) BeforeCreate(*gorm.DB) error               { ensureUUID(&m.ID); return nil }
