package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// customize the base if needed
type Base struct {
	ID        uuid.UUID      `json:"id" gorm:"primary_key;type:char(36);"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

func (b *Base) AssignUUID() {
	b.ID = uuid.New()
}
