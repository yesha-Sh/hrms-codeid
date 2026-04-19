package audit

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"hrms/backend/internal/httpx"
	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/shared"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) Handler {
	return Handler{db: db}
}

func MountRoutes(router chi.Router, handler Handler) {
	router.Get("/audit-logs", handler.list)
}

func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	pagination := shared.ParsePagination(r, "created_at")
	query := h.db.Model(&models.AuditLog{}).Preload("ActorUser")

	if search := r.URL.Query().Get("search"); search != "" {
		term := "%" + search + "%"
		query = query.Joins("LEFT JOIN users actor_users ON actor_users.id = audit_logs.actor_user_id").
			Where("audit_logs.entity ILIKE ? OR audit_logs.action ILIKE ? OR actor_users.email ILIKE ?", term, term, term)
	}
	if entity := r.URL.Query().Get("entity"); entity != "" {
		query = query.Where("entity = ?", entity)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not count audit logs", nil)
		return
	}

	var logs []models.AuditLog
	if err := query.Order("audit_logs." + pagination.Sort + " " + pagination.Order).Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit).Find(&logs).Error; err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list audit logs", nil)
		return
	}

	items := make([]map[string]any, 0, len(logs))
	for _, item := range logs {
		entry := map[string]any{
			"id":         item.ID,
			"action":     item.Action,
			"entity":     item.Entity,
			"entity_id":  item.EntityID,
			"metadata":   item.Metadata,
			"created_at": item.CreatedAt,
		}
		if item.ActorUser != nil {
			entry["actor_user_id"] = item.ActorUser.ID
			entry["actor_email"] = item.ActorUser.Email
		}
		items = append(items, entry)
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"items": items,
		"meta":  map[string]any{"page": pagination.Page, "limit": pagination.Limit, "total": total},
	})
}
