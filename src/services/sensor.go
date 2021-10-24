package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"

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
	newSensorBody := new(CreateSensorBody)
	if err := c.ShouldBindJSON(&newSensorBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	newSensor, err := models.CreateSensor(newSensorBody.Name, newSensorBody.Type)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
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
	sensors, err := models.ListSensors()
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
	idString := c.Param("sensorId")
	uuid, err := uuid.Parse(idString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprint("malformed sensor id", idString)})
	}
	sensor, err := models.GetSensorByID(uuid)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, GetSensorResponse{Msg: "OK", Sensor: *sensor})
	}
}
