package models

import "time"

type SensorReading struct {
	SensorName string    `json:"sensorName" bson:"sensorName"`
	Value      float64   `json:"value" bson:"value"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
}

var collectionName = "sensor-readings"

func CreateSensorReading(Sensor *Sensor, Value float64) {
	sr := SensorReading{SensorName: Sensor.Name, Value: Value, CreatedAt: time.Now().UTC()}
	DB.InsertOne(collectionName, sr)
}
