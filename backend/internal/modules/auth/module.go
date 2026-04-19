package authmodule

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	internalauth "hrms/backend/internal/auth"
	"hrms/backend/internal/config"
	"hrms/backend/internal/httpx"
	"hrms/backend/internal/middleware"
	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/shared"
)

const refreshCookieName = "hrms_refresh_token"

type Service struct {
	db       *gorm.DB
	cfg      config.Config
	tokens   internalauth.TokenManager
	password internalauth.PasswordHasher
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiresAt   time.Time  `json:"expires_at"`
	User        meResponse `json:"user"`
}

type meResponse struct {
	ID       string           `json:"id"`
	Email    string           `json:"email"`
	Role     string           `json:"role"`
	IsActive bool             `json:"is_active"`
	Employee *employeeSummary `json:"employee,omitempty"`
}

type employeeSummary struct {
	ID             string `json:"id"`
	EmployeeCode   string `json:"employee_code"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	FullName       string `json:"full_name"`
	DepartmentName string `json:"department_name,omitempty"`
	JobTitle       string `json:"job_title,omitempty"`
	WorkMode       string `json:"work_mode,omitempty"`
	ManagementScope string `json:"management_scope,omitempty"`
}

func NewService(cfg config.Config, db *gorm.DB, tokens internalauth.TokenManager, password internalauth.PasswordHasher) Service {
	return Service{db: db, cfg: cfg, tokens: tokens, password: password}
}

func Mount(router chi.Router, service Service) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", service.handleLogin)
		r.Post("/refresh", service.handleRefresh)
		r.Post("/logout", service.handleLogout)
		r.With(middleware.Authenticator(service.tokens)).Get("/me", service.handleMe)
	})
}

func (s Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.Email == "" || req.Password == "" {
		httpx.Error(w, http.StatusBadRequest, "email and password are required", map[string]string{
			"email":    "required",
			"password": "required",
		})
		return
	}

	resp, refreshToken, err := s.login(strings.ToLower(req.Email), req.Password, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	s.setRefreshCookie(w, refreshToken)
	httpx.JSON(w, http.StatusOK, resp)
}

func (s Service) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "missing refresh token", nil)
		return
	}

	resp, newToken, err := s.refresh(cookie.Value, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		http.SetCookie(w, expiredRefreshCookie())
		httpx.Error(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	s.setRefreshCookie(w, newToken)
	httpx.JSON(w, http.StatusOK, resp)
}

func (s Service) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err == nil {
		_ = s.logout(cookie.Value)
	}
	http.SetCookie(w, expiredRefreshCookie())
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s Service) handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "missing auth claims", nil)
		return
	}

	resp, err := s.me(claims)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	httpx.JSON(w, http.StatusOK, resp)
}

func (s Service) login(email, password, userAgent, ipAddress string) (authResponse, string, error) {
	var user models.User
	result := s.db.Preload("Role").Limit(1).Find(&user, "email = ?", email)
	if result.Error != nil {
		return authResponse{}, "", errors.New("invalid email or password")
	}
	if result.RowsAffected == 0 {
		return authResponse{}, "", errors.New("invalid email or password")
	}
	if !user.IsActive {
		return authResponse{}, "", errors.New("user account is inactive")
	}

	ok, err := s.password.Verify(password, user.PasswordHash)
	if err != nil || !ok {
		return authResponse{}, "", errors.New("invalid email or password")
	}

	employeeID := user.EmployeeID
	if employee, findErr := shared.FindLinkedEmployee(s.db, user); findErr == nil {
		employeeID = &employee.ID
	}

	accessToken, accessExpiry, err := s.tokens.GenerateAccessToken(user.ID.String(), user.Role.Code, employeeID)
	if err != nil {
		return authResponse{}, "", errors.New("could not create access token")
	}

	refreshTokenID, refreshToken, refreshExpiry, err := s.tokens.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return authResponse{}, "", errors.New("could not create refresh token")
	}
	refreshID, err := uuid.Parse(refreshTokenID)
	if err != nil {
		return authResponse{}, "", errors.New("could not parse refresh token id")
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	if err := s.db.Model(&user).Where("id = ?", user.ID).Update("last_login_at", now).Error; err != nil {
		return authResponse{}, "", errors.New("could not update login timestamp")
	}

	refresh := models.RefreshToken{
		ID:        refreshID,
		UserID:    user.ID,
		TokenHash: internalauth.HashToken(refreshToken),
		ExpiresAt: refreshExpiry,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}
	if err := s.db.Create(&refresh).Error; err != nil {
		return authResponse{}, "", errors.New("could not persist refresh token")
	}

	me, err := s.me(&internalauth.AccessClaims{UserID: user.ID.String(), Role: user.Role.Code})
	if err != nil {
		return authResponse{}, "", err
	}

	return authResponse{
		AccessToken: accessToken,
		ExpiresAt:   accessExpiry,
		User:        me,
	}, refreshToken, nil
}

func (s Service) refresh(rawToken, userAgent, ipAddress string) (authResponse, string, error) {
	claims, err := s.tokens.ParseRefreshToken(rawToken)
	if err != nil {
		return authResponse{}, "", errors.New("invalid refresh token")
	}

	tokenHash := internalauth.HashToken(rawToken)
	var refresh models.RefreshToken
	err = s.db.Where("token_hash = ? AND revoked_at IS NULL", tokenHash).First(&refresh).Error
	if err != nil {
		return authResponse{}, "", errors.New("refresh token not found")
	}
	if time.Now().UTC().After(refresh.ExpiresAt) {
		return authResponse{}, "", errors.New("refresh token expired")
	}

	var user models.User
	if err := s.db.Preload("Role").First(&user, "id = ?", claims.UserID).Error; err != nil {
		return authResponse{}, "", errors.New("user not found")
	}

	employeeID := user.EmployeeID
	if employee, findErr := shared.FindLinkedEmployee(s.db, user); findErr == nil {
		employeeID = &employee.ID
	}

	accessToken, accessExpiry, err := s.tokens.GenerateAccessToken(user.ID.String(), user.Role.Code, employeeID)
	if err != nil {
		return authResponse{}, "", errors.New("could not create access token")
	}

	now := time.Now().UTC()
	if err := s.db.Model(&refresh).Update("revoked_at", now).Error; err != nil {
		return authResponse{}, "", errors.New("could not rotate refresh token")
	}

	nextRefreshTokenID, nextRefreshToken, refreshExpiry, err := s.tokens.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return authResponse{}, "", errors.New("could not create refresh token")
	}
	nextRefreshID, err := uuid.Parse(nextRefreshTokenID)
	if err != nil {
		return authResponse{}, "", errors.New("could not parse refresh token id")
	}

	nextRefresh := models.RefreshToken{
		ID:        nextRefreshID,
		UserID:    user.ID,
		TokenHash: internalauth.HashToken(nextRefreshToken),
		ExpiresAt: refreshExpiry,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}
	if err := s.db.Create(&nextRefresh).Error; err != nil {
		return authResponse{}, "", errors.New("could not persist refresh token")
	}

	me, err := s.me(&internalauth.AccessClaims{UserID: user.ID.String(), Role: user.Role.Code})
	if err != nil {
		return authResponse{}, "", err
	}

	return authResponse{
		AccessToken: accessToken,
		ExpiresAt:   accessExpiry,
		User:        me,
	}, nextRefreshToken, nil
}

func (s Service) logout(rawToken string) error {
	tokenHash := internalauth.HashToken(rawToken)
	return s.db.Model(&models.RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", tokenHash).Update("revoked_at", time.Now().UTC()).Error
}

func (s Service) me(claims *internalauth.AccessClaims) (meResponse, error) {
	var user models.User
	err := s.db.Preload("Role").First(&user, "id = ?", claims.UserID).Error
	if err != nil {
		return meResponse{}, errors.New("user not found")
	}

	response := meResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Role:     user.Role.Code,
		IsActive: user.IsActive,
	}

	employee, err := shared.FindLinkedEmployee(s.db, user)
	if err == nil {
		response.Employee = &employeeSummary{
			ID:              employee.ID.String(),
			EmployeeCode:    employee.EmployeeCode,
			FirstName:       employee.FirstName,
			LastName:        employee.LastName,
			FullName:        shared.FullName(employee.FirstName, employee.LastName),
			DepartmentName:  employee.Department.Name,
			JobTitle:        employee.Job.Title,
			WorkMode:        employee.WorkMode,
			ManagementScope: employee.ManagementScope,
		}
	}

	return response, nil
}

func (s Service) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.cfg.JWTRefreshTTL.Seconds()),
	})
}

func expiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}
