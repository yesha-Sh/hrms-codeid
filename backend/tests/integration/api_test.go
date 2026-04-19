package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/gorm"

	"hrms/backend/internal/config"
	internaldb "hrms/backend/internal/db"
	"hrms/backend/internal/models"
	"hrms/backend/internal/server"
)

type integrationSuite struct {
	rootDir        string
	cfg            config.Config
	db             *gorm.DB
	sqlDB          *sql.DB
	server         *httptest.Server
	adminEmail     string
	adminPassword  string
	managerEmail   string
	managerPassword string
	teamEmail      string
	teamPassword   string
	employeeEmail  string
	employeePassword string
}

var suite *integrationSuite

func TestMain(m *testing.M) {
	var err error
	suite, err = setupSuite()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "integration setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if closeErr := suite.close(); closeErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "integration cleanup failed: %v\n", closeErr)
		if code == 0 {
			code = 1
		}
	}

	os.Exit(code)
}

func TestAuthLoginRefreshAndLogout(t *testing.T) {
	client := suite.newClient(t)

	login := suite.login(t, client, suite.adminEmail, suite.adminPassword)
	if login.AccessToken == "" {
		t.Fatal("expected access token from login")
	}
	if login.User.Role != "admin" {
		t.Fatalf("expected admin role, got %q", login.User.Role)
	}

	meStatus, meBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/auth/me", login.AccessToken, nil)
	if meStatus != http.StatusOK {
		t.Fatalf("expected /auth/me 200, got %d with %v", meStatus, meBody)
	}

	refreshStatus, refreshBody := suite.requestJSON(t, client, http.MethodPost, "/api/v1/auth/refresh", "", nil)
	if refreshStatus != http.StatusOK {
		t.Fatalf("expected /auth/refresh 200, got %d with %v", refreshStatus, refreshBody)
	}
	refreshedAccess := stringValue(refreshBody, "access_token")
	if refreshedAccess == "" {
		t.Fatal("expected non-empty access token from refresh")
	}

	logoutStatus, logoutBody := suite.requestJSON(t, client, http.MethodPost, "/api/v1/auth/logout", "", nil)
	if logoutStatus != http.StatusOK {
		t.Fatalf("expected /auth/logout 200, got %d with %v", logoutStatus, logoutBody)
	}

	refreshAfterLogoutStatus, _ := suite.requestJSON(t, client, http.MethodPost, "/api/v1/auth/refresh", "", nil)
	if refreshAfterLogoutStatus != http.StatusUnauthorized {
		t.Fatalf("expected refresh after logout to fail with 401, got %d", refreshAfterLogoutStatus)
	}
}

func TestAdminEmployeeCRUDAndManagementFilter(t *testing.T) {
	client := suite.newClient(t)
	login := suite.login(t, client, suite.adminEmail, suite.adminPassword)

	filterStatus, filterBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/employees?page=1&limit=100&management_scope=team_manager", login.AccessToken, nil)
	if filterStatus != http.StatusOK {
		t.Fatalf("expected filtered employees 200, got %d with %v", filterStatus, filterBody)
	}
	if total := intValue(nestedMap(filterBody, "meta"), "total"); total != 2 {
		t.Fatalf("expected 2 team managers in seeded data, got %d", total)
	}

	job := suite.findJobByTitle(t, "System Developer")
	department := suite.findDepartmentByName(t, "System Development")
	location := suite.findLocationByName(t, "Jakarta Headquarters")
	employeeType := suite.findEmployeeTypeByCode(t, "full-time")
	employmentStatus := suite.findEmploymentStatusByCode(t, "active")
	manager := suite.findEmployeeByEmail(t, "salma.nuraini@codeid.co.id")

	suffix := strings.ToLower(uuid.NewString()[:8])
	payload := map[string]any{
		"employee_code":          "CID-T-" + strings.ToUpper(suffix[:4]),
		"first_name":             "Test",
		"last_name":              "Automation",
		"email":                  "auto." + suffix + "@codeid.co.id",
		"phone_number":           "+62-811-9090",
		"hire_date":              "2026-04-09",
		"salary":                 15000000,
		"employee_type_id":       employeeType.ID.String(),
		"employment_status_id":   employmentStatus.ID.String(),
		"department_id":          department.ID.String(),
		"job_id":                 job.ID.String(),
		"location_id":            location.ID.String(),
		"work_mode":              "hybrid",
		"management_scope":       "individual_contributor",
		"manager_employee_id":    manager.ID.String(),
	}

	createStatus, createBody := suite.requestJSON(t, client, http.MethodPost, "/api/v1/employees", login.AccessToken, payload)
	if createStatus != http.StatusCreated {
		t.Fatalf("expected employee create 201, got %d with %v", createStatus, createBody)
	}
	createdID := stringValue(createBody, "id")
	if createdID == "" {
		t.Fatal("expected created employee id")
	}

	getStatus, getBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/employees/"+createdID, login.AccessToken, nil)
	if getStatus != http.StatusOK {
		t.Fatalf("expected created employee fetch 200, got %d with %v", getStatus, getBody)
	}
	if stringValue(getBody, "email") != payload["email"] {
		t.Fatalf("expected created employee email %q, got %q", payload["email"], stringValue(getBody, "email"))
	}

	deleteStatus, deleteBody := suite.requestJSON(t, client, http.MethodDelete, "/api/v1/employees/"+createdID, login.AccessToken, nil)
	if deleteStatus != http.StatusOK {
		t.Fatalf("expected employee delete 200, got %d with %v", deleteStatus, deleteBody)
	}
}

func TestEmployeeCannotEditOrDeleteFinalizedLeave(t *testing.T) {
	client := suite.newClient(t)
	login := suite.login(t, client, "keisha.anindya@codeid.co.id", suite.employeePassword)

	listStatus, listBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/leave-requests?page=1&limit=100", login.AccessToken, nil)
	if listStatus != http.StatusOK {
		t.Fatalf("expected leave list 200, got %d with %v", listStatus, listBody)
	}

	item := findLeaveByStatus(itemsOf(listBody), "rejected")
	if item == nil {
		t.Fatal("expected a rejected leave request for the employee")
	}
	leaveID := stringValue(item, "id")
	updatePayload := map[string]any{
		"leave_type_id": stringValue(item, "leave_type_id"),
		"start_date":    stringValue(item, "start_date"),
		"end_date":      stringValue(item, "end_date"),
		"reason":        "Trying to reopen a finalized leave",
		"status":        "pending",
	}

	updateStatus, updateBody := suite.requestJSON(t, client, http.MethodPut, "/api/v1/leave-requests/"+leaveID, login.AccessToken, updatePayload)
	if updateStatus != http.StatusBadRequest {
		t.Fatalf("expected finalized leave edit to fail with 400, got %d with %v", updateStatus, updateBody)
	}

	deleteStatus, deleteBody := suite.requestJSON(t, client, http.MethodDelete, "/api/v1/leave-requests/"+leaveID, login.AccessToken, nil)
	if deleteStatus != http.StatusBadRequest {
		t.Fatalf("expected finalized leave delete to fail with 400, got %d with %v", deleteStatus, deleteBody)
	}
}

func TestManagerApprovalRules(t *testing.T) {
	client := suite.newClient(t)
	login := suite.login(t, client, "salma.nuraini@codeid.co.id", suite.managerPassword)

	listStatus, listBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/leave-requests?page=1&limit=100", login.AccessToken, nil)
	if listStatus != http.StatusOK {
		t.Fatalf("expected manager leave list 200, got %d with %v", listStatus, listBody)
	}

	target := findLeaveForEmployee(itemsOf(listBody), "Nabila Putri", "pending")
	if target == nil {
		t.Fatal("expected pending team leave request for Nabila Putri")
	}

	updatePayload := map[string]any{
		"employee_id":           stringValue(target, "employee_id"),
		"approver_employee_id":  uuid.NewString(),
		"leave_type_id":         stringValue(target, "leave_type_id"),
		"start_date":            stringValue(target, "start_date"),
		"end_date":              stringValue(target, "end_date"),
		"reason":                stringValue(target, "reason"),
		"status":                "approved",
	}

	updateStatus, updateBody := suite.requestJSON(t, client, http.MethodPut, "/api/v1/leave-requests/"+stringValue(target, "id"), login.AccessToken, updatePayload)
	if updateStatus != http.StatusOK {
		t.Fatalf("expected manager approval 200, got %d with %v", updateStatus, updateBody)
	}
	if stringValue(updateBody, "status") != "approved" {
		t.Fatalf("expected approved status, got %q", stringValue(updateBody, "status"))
	}
	if stringValue(updateBody, "approver_name") != "Salma Nuraini" {
		t.Fatalf("expected approver to be the acting manager, got %q", stringValue(updateBody, "approver_name"))
	}

	createPayload := map[string]any{
		"leave_type_id": suite.findLeaveTypeByCode(t, "annual").ID.String(),
		"start_date":    futureDate(45),
		"end_date":      futureDate(46),
		"reason":        "Manager self-service request",
		"status":        "pending",
	}
	createStatus, createBody := suite.requestJSON(t, client, http.MethodPost, "/api/v1/leave-requests", login.AccessToken, createPayload)
	if createStatus != http.StatusCreated {
		t.Fatalf("expected manager self leave create 201, got %d with %v", createStatus, createBody)
	}

	selfUpdatePayload := map[string]any{
		"leave_type_id": stringValue(createBody, "leave_type_id"),
		"start_date":    stringValue(createBody, "start_date"),
		"end_date":      stringValue(createBody, "end_date"),
		"reason":        "Trying to self-approve",
		"status":        "approved",
	}
	selfUpdateStatus, selfUpdateBody := suite.requestJSON(t, client, http.MethodPut, "/api/v1/leave-requests/"+stringValue(createBody, "id"), login.AccessToken, selfUpdatePayload)
	if selfUpdateStatus != http.StatusBadRequest {
		t.Fatalf("expected manager self-approval to fail with 400, got %d with %v", selfUpdateStatus, selfUpdateBody)
	}
}

func TestTeamManagerAttendanceAndTeamMembershipFlows(t *testing.T) {
	client := suite.newClient(t)
	login := suite.login(t, client, suite.teamEmail, suite.teamPassword)

	attendanceStatus, attendanceBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/attendances?page=1&limit=100", login.AccessToken, nil)
	if attendanceStatus != http.StatusOK {
		t.Fatalf("expected team manager attendance list 200, got %d with %v", attendanceStatus, attendanceBody)
	}
	if total := intValue(nestedMap(attendanceBody, "meta"), "total"); total < 1 {
		t.Fatal("expected at least one attendance record in manager scope")
	}

	nabila := suite.findEmployeeByEmail(t, "nabila.putri@codeid.co.id")
	invalidAttendancePayload := map[string]any{
		"employee_id":      nabila.ID.String(),
		"attendance_date":  futureDate(20),
		"status":           "on time",
		"notes":            "Manager should not create attendance for another employee",
	}
	invalidStatus, invalidBody := suite.requestJSON(t, client, http.MethodPost, "/api/v1/attendances", login.AccessToken, invalidAttendancePayload)
	if invalidStatus != http.StatusBadRequest {
		t.Fatalf("expected manager cross-employee attendance create to fail with 400, got %d with %v", invalidStatus, invalidBody)
	}

	teamManager := suite.findEmployeeByEmail(t, suite.teamEmail)
	validAttendancePayload := map[string]any{
		"employee_id":      teamManager.ID.String(),
		"attendance_date":  futureDate(21),
		"status":           "on time",
		"notes":            "Automated manager attendance entry",
	}
	validStatus, validBody := suite.requestJSON(t, client, http.MethodPost, "/api/v1/attendances", login.AccessToken, validAttendancePayload)
	if validStatus != http.StatusCreated {
		t.Fatalf("expected manager self attendance create 201, got %d with %v", validStatus, validBody)
	}

	teamsStatus, teamsBody := suite.requestJSON(t, client, http.MethodGet, "/api/v1/teams", login.AccessToken, nil)
	if teamsStatus != http.StatusOK {
		t.Fatalf("expected teams list 200, got %d with %v", teamsStatus, teamsBody)
	}
	team := findTeamByName(itemsOf(teamsBody), "Client Solutions Squad")
	if team == nil {
		t.Fatal("expected Client Solutions Squad for seeded team manager")
	}

	availableStatus, availableBody := suite.requestJSON(t, client, http.MethodGet, fmt.Sprintf("/api/v1/teams/%s/available-employees", stringValue(team, "id")), login.AccessToken, nil)
	if availableStatus != http.StatusOK {
		t.Fatalf("expected available employees 200, got %d with %v", availableStatus, availableBody)
	}
	availableItems := itemsOf(availableBody)
	if len(availableItems) == 0 {
		t.Fatal("expected at least one available employee for the team")
	}
	availableEmployee := availableItems[0]

	addStatus, addBody := suite.requestJSON(t, client, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/members", stringValue(team, "id")), login.AccessToken, map[string]any{
		"employee_id": stringValue(availableEmployee, "id"),
	})
	if addStatus != http.StatusCreated {
		t.Fatalf("expected add team member 201, got %d with %v", addStatus, addBody)
	}

	removeStatus, removeBody := suite.requestJSON(t, client, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%s/members/%s", stringValue(team, "id"), stringValue(addBody, "id")), login.AccessToken, nil)
	if removeStatus != http.StatusOK {
		t.Fatalf("expected remove team member 200, got %d with %v", removeStatus, removeBody)
	}
}

func TestSecondaryAssignmentOverlapIsRejected(t *testing.T) {
	client := suite.newClient(t)
	login := suite.login(t, client, suite.adminEmail, suite.adminPassword)

	nabila := suite.findEmployeeByEmail(t, "nabila.putri@codeid.co.id")
	smartsourcing := suite.findDepartmentByName(t, "Smartsourcing")
	outsourcingJob := suite.findJobByTitle(t, "Outsourcing Specialist")

	payload := map[string]any{
		"employee_id":               nabila.ID.String(),
		"job_id":                    outsourcingJob.ID.String(),
		"department_id":             smartsourcing.ID.String(),
		"estimated_hours_per_week":  8,
		"start_date":                "2026-01-20",
		"notes":                     "Overlapping assignment attempt",
	}

	status, body := suite.requestJSON(t, client, http.MethodPost, "/api/v1/employee-role-assignments", login.AccessToken, payload)
	if status != http.StatusBadRequest {
		t.Fatalf("expected overlapping assignment create to fail with 400, got %d with %v", status, body)
	}
}

func setupSuite() (*integrationSuite, error) {
	rootDir, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return nil, err
	}
	_ = godotenv.Load(filepath.Join(rootDir, ".env"))

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Config{
			AppEnv:            "development",
			HTTPPort:          "8080",
			FrontendOrigin:    "http://localhost:5173",
			DBHost:            "127.0.0.1",
			DBPort:            "5432",
			DBName:            "hrms",
			DBUser:            "postgres",
			DBPassword:        "postgres",
			DBSSLMode:         "disable",
			JWTAccessSecret:   "integration-access-secret",
			JWTRefreshSecret:  "integration-refresh-secret",
			JWTAccessTTL:      15 * time.Minute,
			JWTRefreshTTL:     168 * time.Hour,
		}
	}

	tempDBName := fmt.Sprintf("hrms_test_%d", time.Now().UnixNano())
	if err := createDatabase(cfg, tempDBName); err != nil {
		return nil, err
	}

	testCfg := cfg
	testCfg.DBName = tempDBName
	testCfg.AdminSeedEmail = "admin@northstar.id"
	testCfg.AdminSeedPassword = "ChangeMe123!"

	env := envForConfig(testCfg)
	if err := runBackendCommand(rootDir, env, "go", "run", "./cmd/migrate", "up"); err != nil {
		_ = dropDatabase(cfg, tempDBName)
		return nil, err
	}
	if err := runBackendCommand(rootDir, env, "go", "run", "./cmd/seed-admin"); err != nil {
		_ = dropDatabase(cfg, tempDBName)
		return nil, err
	}
	if err := runBackendCommand(rootDir, env, "go", "run", "./cmd/seed-demo"); err != nil {
		_ = dropDatabase(cfg, tempDBName)
		return nil, err
	}

	gormDB, err := internaldb.OpenGORM(testCfg)
	if err != nil {
		_ = dropDatabase(cfg, tempDBName)
		return nil, err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		_ = dropDatabase(cfg, tempDBName)
		return nil, err
	}

	app, err := server.New(testCfg, gormDB)
	if err != nil {
		sqlDB.Close()
		_ = dropDatabase(cfg, tempDBName)
		return nil, err
	}

	return &integrationSuite{
		rootDir:          rootDir,
		cfg:              testCfg,
		db:               gormDB,
		sqlDB:            sqlDB,
		server:           httptest.NewServer(app.Router()),
		adminEmail:       testCfg.AdminSeedEmail,
		adminPassword:    testCfg.AdminSeedPassword,
		managerEmail:     "salma.nuraini@codeid.co.id",
		managerPassword:  "Manager123!",
		teamEmail:        "fajar.maulana@codeid.co.id",
		teamPassword:     "Manager123!",
		employeeEmail:    "nabila.putri@codeid.co.id",
		employeePassword: "Employee123!",
	}, nil
}

func (s *integrationSuite) close() error {
	if s.server != nil {
		s.server.Close()
	}
	if s.sqlDB != nil {
		_ = s.sqlDB.Close()
	}
	return dropDatabase(s.cfg, s.cfg.DBName)
}

func (s *integrationSuite) newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &http.Client{Jar: jar, Timeout: 20 * time.Second}
}

func (s *integrationSuite) login(t *testing.T, client *http.Client, email, password string) loginResponse {
	t.Helper()
	status, body := s.requestJSON(t, client, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	if status != http.StatusOK {
		t.Fatalf("login failed for %s: status %d body %v", email, status, body)
	}

	response := loginResponse{
		AccessToken: stringValue(body, "access_token"),
		User: loginUser{
			Role:  stringValue(nestedMap(body, "user"), "role"),
			Email: stringValue(nestedMap(body, "user"), "email"),
		},
	}
	return response
}

func (s *integrationSuite) requestJSON(t *testing.T, client *http.Client, method, path, accessToken string, payload any) (int, map[string]any) {
	t.Helper()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload for %s %s: %v", method, path, err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, s.server.URL+path, body)
	if err != nil {
		t.Fatalf("create request %s %s: %v", method, path, err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("execute request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response %s %s: %v", method, path, err)
	}
	if len(rawBody) == 0 {
		return resp.StatusCode, map[string]any{}
	}

	var decoded map[string]any
	if err := json.Unmarshal(rawBody, &decoded); err != nil {
		t.Fatalf("decode response %s %s: %v body=%s", method, path, err, string(rawBody))
	}
	return resp.StatusCode, decoded
}

func (s *integrationSuite) findEmployeeByEmail(t *testing.T, email string) models.Employee {
	t.Helper()
	var employee models.Employee
	if err := s.db.Where("email = ?", strings.ToLower(email)).First(&employee).Error; err != nil {
		t.Fatalf("find employee by email %s: %v", email, err)
	}
	return employee
}

func (s *integrationSuite) findDepartmentByName(t *testing.T, name string) models.Department {
	t.Helper()
	var department models.Department
	if err := s.db.Where("name = ?", name).First(&department).Error; err != nil {
		t.Fatalf("find department %s: %v", name, err)
	}
	return department
}

func (s *integrationSuite) findJobByTitle(t *testing.T, title string) models.Job {
	t.Helper()
	var job models.Job
	if err := s.db.Where("title = ?", title).First(&job).Error; err != nil {
		t.Fatalf("find job %s: %v", title, err)
	}
	return job
}

func (s *integrationSuite) findLocationByName(t *testing.T, name string) models.Location {
	t.Helper()
	var location models.Location
	if err := s.db.Where("name = ?", name).First(&location).Error; err != nil {
		t.Fatalf("find location %s: %v", name, err)
	}
	return location
}

func (s *integrationSuite) findEmployeeTypeByCode(t *testing.T, code string) models.EmployeeType {
	t.Helper()
	var employeeType models.EmployeeType
	if err := s.db.Where("code = ?", code).First(&employeeType).Error; err != nil {
		t.Fatalf("find employee type %s: %v", code, err)
	}
	return employeeType
}

func (s *integrationSuite) findEmploymentStatusByCode(t *testing.T, code string) models.EmploymentStatus {
	t.Helper()
	var status models.EmploymentStatus
	if err := s.db.Where("code = ?", code).First(&status).Error; err != nil {
		t.Fatalf("find employment status %s: %v", code, err)
	}
	return status
}

func (s *integrationSuite) findLeaveTypeByCode(t *testing.T, code string) models.LeaveType {
	t.Helper()
	var leaveType models.LeaveType
	if err := s.db.Where("code = ?", code).First(&leaveType).Error; err != nil {
		t.Fatalf("find leave type %s: %v", code, err)
	}
	return leaveType
}

func createDatabase(cfg config.Config, name string) error {
	adminDSN := adminDatabaseConfig(cfg).DSN()
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE DATABASE "` + name + `"`)
	return err
}

func dropDatabase(cfg config.Config, name string) error {
	adminDSN := adminDatabaseConfig(cfg).DSN()
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	_, _ = db.Exec(`
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = $1 AND pid <> pg_backend_pid()
`, name)
	_, err = db.Exec(`DROP DATABASE IF EXISTS "` + name + `"`)
	return err
}

func adminDatabaseConfig(cfg config.Config) config.Config {
	adminCfg := cfg
	adminCfg.DBName = "postgres"
	return adminCfg
}

func envForConfig(cfg config.Config) []string {
	base := append([]string{}, os.Environ()...)
	overrides := map[string]string{
		"APP_ENV":            cfg.AppEnv,
		"HTTP_PORT":          cfg.HTTPPort,
		"FRONTEND_ORIGIN":    cfg.FrontendOrigin,
		"DB_HOST":            cfg.DBHost,
		"DB_PORT":            cfg.DBPort,
		"DB_NAME":            cfg.DBName,
		"DB_USER":            cfg.DBUser,
		"DB_PASSWORD":        cfg.DBPassword,
		"DB_SSLMODE":         cfg.DBSSLMode,
		"JWT_ACCESS_SECRET":  cfg.JWTAccessSecret,
		"JWT_REFRESH_SECRET": cfg.JWTRefreshSecret,
		"JWT_ACCESS_TTL":     cfg.JWTAccessTTL.String(),
		"JWT_REFRESH_TTL":    cfg.JWTRefreshTTL.String(),
		"ADMIN_SEED_EMAIL":   cfg.AdminSeedEmail,
		"ADMIN_SEED_PASSWORD": cfg.AdminSeedPassword,
	}

	for key, value := range overrides {
		base = append(base, key+"="+value)
	}
	return base
}

func runBackendCommand(rootDir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = rootDir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}

func itemsOf(body map[string]any) []map[string]any {
	raw, ok := body["items"].([]any)
	if !ok {
		return []map[string]any{}
	}
	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if decoded, ok := item.(map[string]any); ok {
			items = append(items, decoded)
		}
	}
	return items
}

func nestedMap(body map[string]any, key string) map[string]any {
	value, ok := body[key].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return value
}

func stringValue(body map[string]any, key string) string {
	value, ok := body[key]
	if !ok || value == nil {
		return ""
	}
	switch cast := value.(type) {
	case string:
		return cast
	default:
		return fmt.Sprintf("%v", cast)
	}
}

func intValue(body map[string]any, key string) int {
	value, ok := body[key]
	if !ok || value == nil {
		return 0
	}
	switch cast := value.(type) {
	case float64:
		return int(cast)
	case int:
		return cast
	default:
		return 0
	}
}

func findLeaveByStatus(items []map[string]any, status string) map[string]any {
	for _, item := range items {
		if strings.EqualFold(stringValue(item, "status"), status) {
			return item
		}
	}
	return nil
}

func findLeaveForEmployee(items []map[string]any, employeeName, status string) map[string]any {
	for _, item := range items {
		if stringValue(item, "employee_name") == employeeName && strings.EqualFold(stringValue(item, "status"), status) {
			return item
		}
	}
	return nil
}

func findTeamByName(items []map[string]any, name string) map[string]any {
	for _, item := range items {
		if stringValue(item, "name") == name {
			return item
		}
	}
	return nil
}

func futureDate(offsetDays int) string {
	return time.Now().UTC().AddDate(0, 0, offsetDays).Format("2006-01-02")
}

type loginResponse struct {
	AccessToken string
	User        loginUser
}

type loginUser struct {
	Role  string
	Email string
}
