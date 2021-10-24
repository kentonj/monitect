package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SensorReading struct {
	Base
	Value    float64   `json:"value"`
	SensorID uuid.UUID `json:"-"`
	Sensor   Sensor    `json:"-"`
}

func CreateSensorReading(sensorID uuid.UUID, value float64) (s *SensorReading, err error) {
	sensorReading := SensorReading{
		Value:    value,
		SensorID: sensorID,
	}
	if sensorReadingId, err := uuid.NewRandom(); err != nil {
		return nil, err
	} else {
		sensorReading.ID = sensorReadingId
	}
	res := DB.Create(&sensorReading)
	if res.Error != nil {
		return nil, res.Error
	} else {
		return &sensorReading, nil
	}
}

func FindSensorReadings(query *gorm.DB) (s *[]SensorReading, e error) {
	sensorsReadings := make([]SensorReading, 0)
	res := query.Find(&sensorsReadings)
	if res.Error != nil {
		return nil, res.Error
	} else {
		return &sensorsReadings, nil
	}
}

func ListSensorReadings(sensorId uuid.UUID, limit int) (s *[]SensorReading, e error) {
	// get readings between a date range with a configurable limit, will get the results in desc order
	if limit == 0 {
		limit = 1000
	}
	query := DB.Where("sensor_id = ?", sensorId).Order("created_at desc").Limit(limit)
	return FindSensorReadings(query)
}
