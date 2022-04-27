package storage

import (
	"log"
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
	// assign a UUID, throw an error if something goes wrong
	if id, err := uuid.NewRandom(); err != nil {
		log.Fatal("Couldn't generate a new UUID, wtf mate")
	} else {
		b.ID = id
	}
}
