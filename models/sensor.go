package models

import (
	"time"
)

type Sensor struct {
	Name      string    `json:"name" bson:"_id"`
	Type      string    `json:"type" bson:"type"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

func CreateSensor(name string, sensorType string) *Sensor {
	sensor := Sensor{
		Name:      name,
		Type:      sensorType,
		CreatedAt: time.Now().UTC(),
	}
	DB.InsertOne("sensors", sensor)
	return &sensor
}

func GetSensorByName(name string) *Sensor {
	// sensor name is stored as the id
	sensor := new(Sensor)
	res := DB.FindById("sensors", name)
	res.Decode(&sensor)
	return sensor
}
