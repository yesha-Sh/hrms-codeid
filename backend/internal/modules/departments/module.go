package departments

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
	Name               string  `json:"name"`
	ParentDepartmentID *string `json:"parent_department_id"`
	ManagerEmployeeID  *string `json:"manager_employee_id"`
	LocationID         string  `json:"location_id"`
	Level              int     `json:"level"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/departments", func(r chi.Router) {
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
	query := s.db.Model(&models.Department{}).
		Preload("Location").
		Preload("ManagerEmployee").
		Preload("ParentDepartment").
		Where("departments.organization_id = ?", organizationID)

	if search := r.URL.Query().Get("search"); search != "" {
		term := "%" + search + "%"
		query = query.Joins("LEFT JOIN employees manager_employees ON manager_employees.id = departments.manager_employee_id").
			Where("departments.name ILIKE ? OR manager_employees.first_name ILIKE ? OR manager_employees.last_name ILIKE ?", term, term, term)
	}

	if managerID := r.URL.Query().Get("manager_id"); managerID != "" {
		query = query.Where("manager_employee_id = ?", managerID)
	}
	if locationID := r.URL.Query().Get("location_id"); locationID != "" {
		query = query.Where("location_id = ?", locationID)
	}
	if level := r.URL.Query().Get("level"); level != "" {
		query = query.Where("level = ?", level)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count departments", nil)
		return
	}

	var departments []models.Department
	if err := query.Order("departments." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&departments).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list departments", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": toListResponse(departments), "meta": map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total}})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid department id", nil)
		return
	}

	department, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "department not found", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, toResponse(department))
}

func (s Service) create(w http.ResponseWriter, r *http.Request) {
	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	department, err := s.toModel(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := s.db.Create(&department).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create department", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "department", &department.ID, map[string]any{"name": department.Name})

	fresh, _ := s.findByID(department.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid department id", nil)
		return
	}

	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	var department models.Department
	if err := s.db.First(&department, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusNotFound, "department not found", nil)
		return
	}

	updated, err := s.toModel(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updated.ID = department.ID
	updated.OrganizationID = department.OrganizationID
	if err := s.db.Model(&department).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update department", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "department", &department.ID, map[string]any{"name": updated.Name})

	fresh, _ := s.findByID(id)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid department id", nil)
		return
	}

	if err := s.db.Delete(&models.Department{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete department", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "department", &id, nil)

	httpx.JSON(w, http.StatusOK, map[string]string{"message": "department deleted"})
}

func (s Service) findByID(id uuid.UUID) (models.Department, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Department{}, err
	}
	var department models.Department
	err = s.db.Preload("Location").Preload("ManagerEmployee").Preload("ParentDepartment").First(&department, "id = ? AND organization_id = ?", id, organizationID).Error
	return department, err
}

func (s Service) toModel(req upsertRequest) (models.Department, error) {
	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		return models.Department{}, err
	}
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Department{}, err
	}

	department := models.Department{Name: req.Name, LocationID: locationID, Level: req.Level, OrganizationID: organizationID}

	if req.ParentDepartmentID != nil && *req.ParentDepartmentID != "" {
		parentID, err := uuid.Parse(*req.ParentDepartmentID)
		if err != nil {
			return models.Department{}, err
		}
		department.ParentDepartmentID = &parentID
	}

	if req.ManagerEmployeeID != nil && *req.ManagerEmployeeID != "" {
		managerID, err := uuid.Parse(*req.ManagerEmployeeID)
		if err != nil {
			return models.Department{}, err
		}
		department.ManagerEmployeeID = &managerID
	}

	return department, nil
}

func toListResponse(items []models.Department) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.Department) map[string]any {
	response := map[string]any{"id": item.ID, "organization_id": item.OrganizationID, "name": item.Name, "location_id": item.LocationID, "location_name": item.Location.Name, "level": item.Level, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
	if item.ParentDepartment != nil {
		response["parent_department_id"] = item.ParentDepartment.ID
		response["parent_department_name"] = item.ParentDepartment.Name
	}
	if item.ManagerEmployee != nil {
		response["manager_employee_id"] = item.ManagerEmployee.ID
		response["manager_name"] = shared.FullName(item.ManagerEmployee.FirstName, item.ManagerEmployee.LastName)
	}
	return response
}
