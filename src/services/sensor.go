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

type CreateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CreateSensorResponse struct {
	Msg    string        `json:"msg"`
	Sensor models.Sensor `json:"sensor"`
}

func CreateSensor(c *gin.Context) {
	// register a sensor, error if it already exists
	newSensorBody := new(CreateSensorBody)
	if err := c.ShouldBindJSON(&newSensorBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	existingSensor, err := models.GetSensorByName(newSensorBody.Name)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if existingSensor != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("sensor with name %s already exists", newSensorBody.Name)})
		return
	}
	newSensor, err := models.CreateSensor(newSensorBody.Name, newSensorBody.Type)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusCreated, CreateSensorResponse{Msg: "Successfully created sensor", Sensor: *newSensor})
	}
}

type ListSensorsResponse struct {
	Msg     string          `json:"msg"`
	Sensors []models.Sensor `json:"sensors"`
	Count   int             `json:"count"`
}

func ListSensors(c *gin.Context) {
	// list sensors
	listOpts := options.Find().SetSort(bson.D{{"_id", -1}}).SetLimit(10)
	queryParams := c.Request.URL.Query()
	fmt.Println(queryParams)
	filter := ParamsToFilter(queryParams)
	fmt.Println("using the following filter")
	fmt.Println(filter)
	sensors, err := models.ListSensors(filter, listOpts)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, ListSensorsResponse{Msg: "OK", Sensors: *sensors, Count: len(*sensors)})
	}
}

type GetSensorResponse struct {
	Msg    string        `json:"msg"`
	Sensor models.Sensor `json:"sensor"`
}

func GetSensor(c *gin.Context) {
	id := c.Param("sensorId")
	sensor, err := models.GetSensorById(id)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, GetSensorResponse{Msg: "OK", Sensor: *sensor})
	}
}
