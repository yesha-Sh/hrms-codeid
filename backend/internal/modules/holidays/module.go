package holidays

import (
	"net/http"
	"time"

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
	Name        string  `json:"name"`
	HolidayDate string  `json:"holiday_date"`
	LocationID  *string `json:"location_id"`
	Notes       string  `json:"notes"`
}

func NewService(db *gorm.DB, audit audit.Service) Service {
	return Service{db: db, audit: audit}
}

func Mount(router chi.Router, service Service) {
	router.Route("/holidays", func(r chi.Router) {
		r.Get("/", service.list)
		r.Post("/", service.create)
		r.Get("/{id}", service.get)
		r.Put("/{id}", service.update)
		r.Delete("/{id}", service.delete)
	})
}

func (s Service) list(w http.ResponseWriter, r *http.Request) {
	pagination := shared.ParsePagination(r, "holiday_date")
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "organization not found", nil)
		return
	}

	query := s.db.Model(&models.Holiday{}).
		Where("holidays.organization_id = ?", organizationID).
		Preload("Location")
	if search := r.URL.Query().Get("search"); search != "" {
		term := "%" + search + "%"
		query = query.Joins("LEFT JOIN locations ON locations.id = holidays.location_id").
			Where("holidays.name ILIKE ? OR locations.name ILIKE ?", term, term)
	}
	if year := r.URL.Query().Get("year"); year != "" {
		query = query.Where("EXTRACT(YEAR FROM holiday_date) = ?", year)
	}
	if locationID := r.URL.Query().Get("location_id"); locationID != "" {
		query = query.Where("location_id = ?", locationID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count holidays", nil)
		return
	}

	var items []models.Holiday
	if err := query.Order("holidays." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&items).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list holidays", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"items": toListResponse(items), "meta": map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total}})
}

func (s Service) get(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid holiday id", nil)
		return
	}

	item, err := s.findByID(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "holiday not found", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, toResponse(item))
}

func (s Service) create(w http.ResponseWriter, r *http.Request) {
	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	item, err := s.toModel(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := s.db.Create(&item).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create holiday", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "create", "holiday", &item.ID, map[string]any{"name": item.Name})

	fresh, _ := s.findByID(item.ID)
	httpx.JSON(w, http.StatusCreated, toResponse(fresh))
}

func (s Service) update(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid holiday id", nil)
		return
	}

	var current models.Holiday
	if err := s.db.First(&current, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusNotFound, "holiday not found", nil)
		return
	}

	var req upsertRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updated, err := s.toModel(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	updated.ID = current.ID
	updated.OrganizationID = current.OrganizationID

	if err := s.db.Model(&current).Updates(updated).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not update holiday", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "update", "holiday", &current.ID, map[string]any{"name": updated.Name})

	fresh, _ := s.findByID(current.ID)
	httpx.JSON(w, http.StatusOK, toResponse(fresh))
}

func (s Service) delete(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid holiday id", nil)
		return
	}

	if err := s.db.Delete(&models.Holiday{}, "id = ?", id).Error; err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not delete holiday", nil)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	actorID, _ := shared.CurrentUserUUID(claims)
	_ = s.audit.Log(actorID, "delete", "holiday", &id, nil)

	httpx.JSON(w, http.StatusOK, map[string]string{"message": "holiday deleted"})
}

func (s Service) findByID(id uuid.UUID) (models.Holiday, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Holiday{}, err
	}
	var item models.Holiday
	err = s.db.Preload("Location").First(&item, "id = ? AND organization_id = ?", id, organizationID).Error
	return item, err
}

func (s Service) toModel(req upsertRequest) (models.Holiday, error) {
	organizationID, err := shared.MustOrganizationID(s.db)
	if err != nil {
		return models.Holiday{}, err
	}
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		return models.Holiday{}, err
	}
	var locationID *uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		parsed, err := shared.ParseUUID(*req.LocationID)
		if err != nil {
			return models.Holiday{}, err
		}
		locationID = &parsed
	}

	return models.Holiday{OrganizationID: organizationID, Name: req.Name, HolidayDate: holidayDate, LocationID: locationID, Notes: req.Notes}, nil
}

func toListResponse(items []models.Holiday) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item models.Holiday) map[string]any {
	response := map[string]any{
		"id":              item.ID,
		"organization_id": item.OrganizationID,
		"name":            item.Name,
		"holiday_date":    item.HolidayDate.Format("2006-01-02"),
		"notes":           item.Notes,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	}
	if item.LocationID != nil {
		response["location_id"] = item.LocationID
	}
	if item.Location != nil {
		response["location_name"] = item.Location.Name
	}
	response["year"] = item.HolidayDate.Year()
	return response
}
