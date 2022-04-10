package models

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

type Sensor struct {
	Base
	Name string `json:"name" gorm:"unique"`
	Type string `json:"type"`
}

func (s *Sensor) Update(body *UpdateSensorBody) {
	// update the sensor with the update sensor body
	if body.Name != nil {
		s.Name = *body.Name
	}
	if body.Type != nil {
		s.Type = *body.Type
	}
}

func CreateSensor(createSensorBody *CreateSensorBody) (s *Sensor, err error) {
	if createSensorBody.Name == nil {
		return nil, errors.New("sensor name required")
	}
	sensor := Sensor{
		Name: *createSensorBody.Name,
		Type: *createSensorBody.Type,
	}
	sensor.AssignUUID()
	res := DB.Create(&sensor)
	if res.Error != nil {
		log.Printf("unable to insert record %s: %v", sensor.Name, res.Error)
		return nil, res.Error
	} else {
		log.Printf("Created sensor with id %d", sensor.ID)
	}
	return &sensor, nil
}

func UpdateSensor(id uuid.UUID, updateSensorBody *UpdateSensorBody) error {
	// update a monitor with non-null values provided in the update monitor body
	var sensor Sensor
	if res := DB.First(&sensor, id); res.Error != nil {
		return res.Error
	}
	sensor.Update(updateSensorBody)
	if res := DB.Save(&sensor); res.Error != nil {
		return res.Error
	} else {
		return nil
	}
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
	if res := DB.Find(&sensors); res.Error != nil {
		return nil, res.Error
	} else {
		return &sensors, nil
	}
}
