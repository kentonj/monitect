package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kentonj/monitect/src/models"
	"github.com/mattn/go-sqlite3"
)

func CreateSensor(c *gin.Context) {
	var newSensorBody models.CreateSensorBody
	if err := c.ShouldBindJSON(&newSensorBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	newSensor, err := models.CreateSensor(&newSensorBody)
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
		c.JSON(http.StatusCreated, models.CreateSensorResponse{BaseResponse: models.SUCCESS, Sensor: *newSensor})
	}
}

func UpdateSensor(c *gin.Context) {
	idString := c.Param("sensorId")
	id, err := uuid.Parse(idString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprint("malformed sensor id ", idString)})
	} else {
		log.Print(id)
	}
	updateSensorBody := new(models.UpdateSensorBody)
	if err := c.ShouldBindJSON(&updateSensorBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "unable to parse the request body"})
	} else {
		log.Print(*updateSensorBody.Name)
	}
	if err := models.UpdateSensor(id, updateSensorBody); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "unable to update the specified sensor"})
	}
	// } else {
	// 	c.JSON(http.StatusOK, models.SUCCESS)
	// }
	c.JSON(http.StatusOK, models.SUCCESS)
}

func ListSensors(c *gin.Context) {
	// list sensors
	sensors, err := models.ListSensors()
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, models.ListSensorsResponse{Msg: "OK", Sensors: *sensors, Count: len(*sensors)})
	}
}

func GetSensor(c *gin.Context) {
	idString := c.Param("sensorId")
	uuid, err := uuid.Parse(idString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprint("malformed sensor id ", idString)})
	}
	sensor, err := models.GetSensorByID(uuid)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, models.GetSensorResponse{Msg: "OK", Sensor: *sensor})
	}
}
