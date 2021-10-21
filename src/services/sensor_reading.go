package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kentonj/monitect/src/models"
)

type CreateSensorReadingBody struct {
	Value float64 `json:"value"`
}

type CreateSensorReadingResponse struct {
	Msg           string               `json:"msg"`
	SensorReading models.SensorReading `json:"sensorReading"`
}

func CreateSensorReading(c *gin.Context) {
	newSensorReadingBody := new(CreateSensorReadingBody)
	if err := c.ShouldBindJSON(&newSensorReadingBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// verify that there is a sensor with this id
	sensorId := c.Param("sensorId")
	if sensor, err := models.GetSensorById(sensorId); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if sensor == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("sensor with id %s not found", sensorId)})
		return
	}
	// now we can safely write a sensor reading
	newSensorReading, err := models.CreateSensorReading(sensorId, newSensorReadingBody.Value)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusCreated, CreateSensorReadingResponse{Msg: "Successfully created sensor reading", SensorReading: *newSensorReading})
	}
}

type ListSensorReadingsResponse struct {
	Msg            string                 `json:"msg"`
	SensorReadings []models.SensorReading `json:"sensorsReadings"`
	Count          int                    `json:"count"`
}

func ListSensorReadings(c *gin.Context) {
	// list sensors
	listOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(100)
	sensorReadings, err := models.ListSensorReadings(bson.D{}, listOpts)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, ListSensorReadingsResponse{Msg: "OK", SensorReadings: *sensorReadings, Count: len(*sensorReadings)})
	}
}
