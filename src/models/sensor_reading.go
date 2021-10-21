package models

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const sensorReadingCollection = "sensor-readings"

type SensorReading struct {
	ID        interface{} `json:"id" bson:"_id,omitempty"`
	SensorId  string      `json:"sensorId" bson:"sensorId"`
	Value     float64     `json:"value" bson:"value"`
	CreatedAt time.Time   `json:"createdAt" bson:"createdAt"`
}

func CreateSensorReading(sensorId string, value float64) (s *SensorReading, err error) {
	sensorReading := SensorReading{
		SensorId:  sensorId,
		Value:     value,
		CreatedAt: time.Now().UTC(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	collection := DB.Collection(sensorReadingCollection)
	res, err := collection.InsertOne(ctx, sensorReading)
	if err != nil {
		return nil, errors.New("unable to insert record")
	} else {
		log.Printf("inserted record: collection %s object id %s", sensorReadingCollection, res.InsertedID)
	}
	sensorReading.ID = res.InsertedID
	return &sensorReading, nil
}

func ListSensorReadings(filter interface{}, opts *options.FindOptions) (s *[]SensorReading, e error) {
	// default options
	if opts == nil {
		opts = options.Find().SetSort(bson.D{{"createdAt", -1}}).SetLimit(1000)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	cur, err := DB.Collection(sensorReadingCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.New("unable to decode cursor results into sensors")
	}
	defer cur.Close(ctx)
	// empty list
	sensorReadings := make([]SensorReading, 0)
	if err := cur.All(ctx, &sensorReadings); err != nil {
		return nil, errors.New("unable to decode cursor results into sensors")
	}
	return &sensorReadings, nil
}
