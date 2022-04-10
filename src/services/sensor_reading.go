package services

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"

	"github.com/kentonj/monitect/src/models"
)

func CreateSensorReading(c *gin.Context) {
	// verify that there is a sensor with this id
	sensorIdString := c.Param("sensorId")
	sensorId, err := uuid.Parse(sensorIdString)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprint("malformed sensor id ", sensorIdString)})
		return
	}
	var newSensorReadingBody models.CreateSensorReadingBody
	if err := c.ShouldBindJSON(&newSensorReadingBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	newSensorReading, err := models.CreateSensorReading(sensorId, &newSensorReadingBody)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "specified sensor id does not exist"})
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusCreated, models.CreateSensorReadingResponse{Msg: "Successfully created sensor reading", SensorReading: *newSensorReading})
	}
}

func ListSensorReadings(c *gin.Context) {
	// list sensors
	sensorIdString := c.Param("sensorId")
	sensorId, err := uuid.Parse(sensorIdString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprint("malformed sensor id ", sensorIdString)})
		return
	}
	limitString := c.Query("limit")
	var limit int
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprint("invalid limit ", limitString)})
			return
		}
	} else {
		limit = 1000
	}
	sensorReadings, err := models.ListSensorReadings(sensorId, limit)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, models.ListSensorReadingsResponse{Msg: "OK", SensorReadings: *sensorReadings, Count: len(*sensorReadings)})
	}
}
