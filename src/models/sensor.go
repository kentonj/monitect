package models

import (
	"log"

	"github.com/google/uuid"
)

type Sensor struct {
	Base
	Name string `json:"name" gorm:"unique"`
	Type string `json:"type"`
}

func CreateSensor(name string, sensorType string) (s *Sensor, err error) {
	sensor := Sensor{
		Name: name,
		Type: sensorType,
	}
	if sensorId, err := uuid.NewRandom(); err != nil {
		return nil, err
	} else {
		sensor.ID = sensorId
	}
	res := DB.Create(&sensor)
	if res.Error != nil {
		log.Printf("unable to insert record %s: %v", sensor.Name, res.Error)
		return nil, res.Error
	} else {
		log.Printf("Created sensor with id %d", sensor.ID)
	}
	return &sensor, nil
}

func GetSensorByName(name string) (s *Sensor, e error) {
	// find a sensor by name, return nil if it doesn't exist
	var sensor Sensor
	res := DB.First(&sensor, "name = ?", name)
	if res.Error != nil {
		return nil, res.Error
	} else {
		return &sensor, nil
	}
}

func GetSensorByID(uuid uuid.UUID) (s *Sensor, e error) {
	// find a sensor by id, return nil if it doesn't exist
	var sensor Sensor
	res := DB.First(&sensor, uuid)
	if res.Error != nil {
		return nil, res.Error
	} else {
		return &sensor, nil
	}
}

func ListSensors() (s *[]Sensor, e error) {
	// empty list
	sensors := make([]Sensor, 0)
	// LIKE
	res := DB.Find(&sensors)
	if res.Error != nil {
		return nil, res.Error
	} else {
		return &sensors, nil
	}
}
