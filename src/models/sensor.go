package models

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const sensorCollection string = "sensors"

type Sensor struct {
	ID        interface{} `json:"id" bson:"_id,omitempty"`
	Name      string      `json:"name" bson:"name"`
	Type      string      `json:"type" bson:"type"`
	CreatedAt time.Time   `json:"createdAt" bson:"createdAt"`
}

func CreateSensor(name string, sensorType string) (s *Sensor, err error) {
	sensor := Sensor{
		Name:      name,
		Type:      sensorType,
		CreatedAt: time.Now().UTC(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	collection := DB.Collection(sensorCollection)
	res, err := collection.InsertOne(ctx, sensor)
	if err != nil {
		log.Printf("unable to insert record %s", sensor.Name)
		return nil, errors.New("unable to insert record")
	} else {
		log.Printf("inserted record: collection %s object id %s", sensorCollection, res.InsertedID)
	}
	sensor.ID = res.InsertedID
	return &sensor, nil
}

func GetSensorByFilter(filter interface{}) (s *Sensor, e error) {
	// find a sensor by name, return nil if it doesn't exist
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	collection := DB.Collection(sensorCollection)
	res := collection.FindOne(ctx, filter)
	findErr := res.Err()
	if findErr == mongo.ErrNoDocuments {
		// not an error, consider changing this to be more clear later
		return nil, nil
	} else if findErr != nil {
		return nil, findErr
	}
	sensor := new(Sensor)
	res.Decode(&sensor)
	return sensor, nil
}

func GetSensorByName(name string) (s *Sensor, e error) {
	// find a sensor by name, return nil if it doesn't exist
	sensor, err := GetSensorByFilter(bson.M{"name": name})
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

func GetSensorById(id string) (s *Sensor, e error) {
	// find a sensor by name, return nil if it doesn't exist
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("provided hex is not a valid objectID")
	}
	return GetSensorByFilter(bson.M{"_id": objectId})
}

func ListSensors(filter interface{}, opts *options.FindOptions) (s *[]Sensor, e error) {
	// default options
	if opts == nil {
		opts = options.Find().SetSort(bson.D{{"createdAt", -1}}).SetLimit(10)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	cur, err := DB.Collection(sensorCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.New("unable to decode cursor results into sensors")
	}
	defer cur.Close(ctx)
	// empty list
	sensors := make([]Sensor, 0)
	if err := cur.All(ctx, &sensors); err != nil {
		return nil, errors.New("unable to decode cursor results into sensors")
	}
	return &sensors, nil
}
