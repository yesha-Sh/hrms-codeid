package jobs

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

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

type upsertRequest struct {
	Title               string   `json:"title"`
	PrimaryDepartmentID string   `json:"primary_department_id"`
	JobLevelID          string   `json:"job_level_id"`
	MinSalary           *float64 `json:"min_salary"`
	MaxSalary           *float64 `json:"max_salary"`
	JobDescription      string   `json:"job_description"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/jobs", func(r chi.Router) {
		r.Get("/", service.list)
		r.Post("/", service.create)
		r.Get("/{id}", service.get)
		r.Put("/{id}", service.update)
		r.Delete("/{id}", service.delete)
	})
}

func (s Service) list(w http.ResponseWriter, r *http.Request) {
	pagination := shared.ParsePagination(r, "created_at")
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}
	query := s.db.Model(&models.Job{}).
		Preload("PrimaryDepartment").
		Preload("JobLevel").
		Where("organization_id = ?", organizationID)
	if search := r.URL.Query().Get("search"); search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("primary_department_id = ?", departmentID)
	}
	if levelID := r.URL.Query().Get("job_level_id"); levelID != "" {
		query = query.Where("job_level_id = ?", levelID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count jobs", nil)
		return
	}

	var jobs []models.Job
	if err := query.Order("jobs." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&jobs).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list jobs", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": toListResponse(jobs), "meta": map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total}})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid job id", nil)
		return
	}

	job, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "job not found", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, toResponse(job))
}

func (s Service) create(w http.ResponseWriter, r *http.Request) {
	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	job, err := s.toModel(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := s.db.Create(&job).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create job", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "job", &job.ID, map[string]any{"title": job.Title})

	fresh, _ := s.findByID(job.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid job id", nil)
		return
	}

	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	var job models.Job
	if err := s.db.First(&job, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusNotFound, "job not found", nil)
		return
	}

	updated, err := s.toModel(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	updated.ID = job.ID
	updated.OrganizationID = job.OrganizationID
	if err := s.db.Model(&job).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update job", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "job", &job.ID, map[string]any{"title": req.Title})

	fresh, _ := s.findByID(id)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid job id", nil)
		return
	}

	if err := s.db.Delete(&models.Job{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete job", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "job", &id, nil)

	httpx.JSON(w, http.StatusOK, map[string]string{"message": "job deleted"})
}

func (s Service) findByID(id uuid.UUID) (models.Job, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Job{}, err
	}
	var job models.Job
	err = s.db.Preload("PrimaryDepartment").Preload("JobLevel").First(&job, "id = ? AND organization_id = ?", id, organizationID).Error
	return job, err
}

func (s Service) toModel(req upsertRequest) (models.Job, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Job{}, err
	}
	departmentID, err := shared.ParseUUID(req.PrimaryDepartmentID)
	if err != nil {
		return models.Job{}, err
	}
	jobLevelID, err := shared.ParseUUID(req.JobLevelID)
	if err != nil {
		return models.Job{}, err
	}
	return models.Job{OrganizationID: organizationID, Title: req.Title, PrimaryDepartmentID: departmentID, JobLevelID: jobLevelID, MinSalary: req.MinSalary, MaxSalary: req.MaxSalary, JobDescription: req.JobDescription}, nil
}

func toListResponse(items []models.Job) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.Job) map[string]any {
	return map[string]any{"id": item.ID, "organization_id": item.OrganizationID, "title": item.Title, "primary_department_id": item.PrimaryDepartmentID, "primary_department_name": item.PrimaryDepartment.Name, "job_level_id": item.JobLevelID, "job_level_name": item.JobLevel.Name, "min_salary": item.MinSalary, "max_salary": item.MaxSalary, "job_description": item.JobDescription, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
}
