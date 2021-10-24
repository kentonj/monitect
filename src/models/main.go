package models

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// globally available mongodb variable for access within each model
var DB *gorm.DB

// customize the base if needed
type Base struct {
	ID        uuid.UUID      `json:"id" gorm:"primary_key;type:char(36);"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

func MigrateSchemas() {
	// migrate all schemas
	DB.AutoMigrate(&Sensor{}, &SensorReading{})
	DB.Migrator().CreateConstraint(&SensorReading{}, "Sensor")
	DB.Migrator().CreateConstraint(&SensorReading{}, "fk_sensor_readings_sensors")
}
