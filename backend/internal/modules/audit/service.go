package audit

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"hrms/backend/internal/models"
	"hrms/backend/internal/modules/shared"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return Service{db: db}
}

func (s Service) Log(actorUserID *uuid.UUID, action, entity string, entityID *uuid.UUID, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}

	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	entry := models.AuditLog{
		OrganizationID: organizationID(s.db),
		ActorUserID:    actorUserID,
		Action:         action,
		Entity:         entity,
		EntityID:       entityID,
		Metadata:       datatypes.JSON(encoded),
	}

	return s.db.Create(&entry).Error
}

func organizationID(db *gorm.DB) uuid.UUID {
	id, err := shared.MustOrganizationID(db)
	if err != nil {
		return uuid.Nil
	}
	return id
}
