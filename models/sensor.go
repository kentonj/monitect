package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Sensor struct {
	Name      string    `json:"name" bson:"name"`
	Type      string    `json:"type" bson:"type"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

func CreateSensor(name string, sensorType string) *mongo.InsertOneResult {
	sensor := Sensor{
		Name:      name,
		Type:      sensorType,
		CreatedAt: time.Now().UTC(),
	}
	return DB.InsertOne("sensors", sensor)
}

func GetSensorByName(name string) *Sensor {
	// find a sensor by name, return nil if it doesn't exist
	res := DB.FindOneByQuery("sensors", bson.M{"name": name})
	if res != nil {
		sensor := new(Sensor)
		res.Decode(&sensor)
		return sensor
	} else {
		return nil
	}
}
