package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"

	internalauth "hrms/backend/internal/auth"
	"hrms/backend/internal/config"
	authmw "hrms/backend/internal/middleware"
	assignmentsmodule "hrms/backend/internal/modules/assignments"
	attendancesmodule "hrms/backend/internal/modules/attendances"
	audithandler "hrms/backend/internal/modules/audit"
	auditmodule "hrms/backend/internal/modules/audit"
	authmodule "hrms/backend/internal/modules/auth"
	dashboardmodule "hrms/backend/internal/modules/dashboard"
	departmentsmodule "hrms/backend/internal/modules/departments"
	employeesmodule "hrms/backend/internal/modules/employees"
	holidaysmodule "hrms/backend/internal/modules/holidays"
	jobsmodule "hrms/backend/internal/modules/jobs"
	leavemodule "hrms/backend/internal/modules/leave"
	lookupsmodule "hrms/backend/internal/modules/lookups"
	profilemodule "hrms/backend/internal/modules/profile"
	teamsmodule "hrms/backend/internal/modules/teams"
)

type Server struct {
	cfg      config.Config
	db       *gorm.DB
	router   chi.Router
	tokens   internalauth.TokenManager
	password internalauth.PasswordHasher
	audit    auditmodule.Service
}

func New(cfg config.Config, db *gorm.DB) (*Server, error) {
	s := &Server{
		cfg:      cfg,
		db:       db,
		tokens:   internalauth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL),
		password: internalauth.NewPasswordHasher(),
		audit:    auditmodule.NewService(db),
	}

	s.router = s.buildRouter()
	return s, nil
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) buildRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.cfg.FrontendOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	authService := authmodule.NewService(s.cfg, s.db, s.tokens, s.password)
	departmentService := departmentsmodule.NewService(s.db, s.audit)
	jobService := jobsmodule.NewService(s.db, s.audit)
	employeeService := employeesmodule.NewService(s.db, s.audit)
	attendanceService := attendancesmodule.NewService(s.db, s.audit)
	leaveService := leavemodule.NewService(s.db, s.audit)
	assignmentService := assignmentsmodule.NewService(s.db, s.audit)
	holidayService := holidaysmodule.NewService(s.db, s.audit)
	lookupService := lookupsmodule.NewService(s.db)
	teamService := teamsmodule.NewService(s.db, s.audit)
	auditHandler := audithandler.NewHandler(s.db)
	dashboardService := dashboardmodule.NewService(s.db)
	profileService := profilemodule.NewService(s.db, s.audit)

	router.Route("/api/v1", func(r chi.Router) {
		authmodule.Mount(r, authService)

		r.Group(func(private chi.Router) {
			private.Use(authmw.Authenticator(s.tokens))
			profilemodule.Mount(private, profileService)
			dashboardmodule.Mount(private, dashboardService)
			lookupsmodule.Mount(private, lookupService)
			teamsmodule.Mount(private, teamService)
			employeesmodule.Mount(private, employeeService)
			attendancesmodule.Mount(private, attendanceService)
			leavemodule.Mount(private, leaveService)
			assignmentsmodule.Mount(private, assignmentService)

			private.Group(func(admin chi.Router) {
				admin.Use(authmw.RequireRoles("admin"))
				departmentsmodule.Mount(admin, departmentService)
				jobsmodule.Mount(admin, jobService)
				holidaysmodule.Mount(admin, holidayService)
				audithandler.MountRoutes(admin, auditHandler)
			})
		})
	})

	return router
}
