package sensorreading

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/storage"
	"gorm.io/gorm"
)

type SensorReading struct {
	storage.Base
	Value    *float64      `json:"value"`
	SensorID uuid.UUID     `json:"-"`
	Sensor   sensor.Sensor `json:"-"`
}

type SensorReadingClient struct {
	db *gorm.DB
}

// return a new sensor reading client
func NewSensorReadingClient(db *gorm.DB) *SensorReadingClient {
	if err := db.AutoMigrate(&SensorReading{}); err != nil {
		log.Fatal("unable to migrate SensorReading object")
	}
	db.Migrator().CreateConstraint(&SensorReading{}, "Sensor")
	db.Migrator().CreateConstraint(&SensorReading{}, "fk_sensor_readings_sensors")
	return &SensorReadingClient{
		db: db,
	}
}

type CreateSensorReadingResponse struct {
	SensorReading *SensorReading `json:"sensorReading"`
}

func (client *SensorReadingClient) CreateSensorReading(c *gin.Context) {
	sensorId, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "unable to parse uuid"})
		return
	}
	var sensorReading SensorReading
	if err := c.ShouldBindJSON(&sensorReading); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "bad json input"})
		return
	}
	if sensorReading.Value == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "sensor reading value cannot be nil"})
		return
	}
	sensorReading.SensorID = sensorId
	sensorReading.AssignUUID()
	if res := client.db.Create(&sensorReading); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "whoops, something went wrong", "details": res.Error})
		return
	} else {
		c.JSON(http.StatusOK, CreateSensorReadingResponse{SensorReading: &sensorReading})
	}
}

type ListSensorReadingsResponse struct {
	SensorReadings *[]SensorReading `json:"sensorReadings"`
	Count          int              `json:"count"`
}

func (client *SensorReadingClient) ListSensorReadings(c *gin.Context) {
	// get readings between a date range with a configurable limit, will get the results in desc order
	sensorId, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad uuid"})
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
	query := client.db.Where("sensor_id = ?", sensorId).Order("created_at desc").Limit(limit)
	sensorReadings := make([]SensorReading, 0)
	if res := query.Find(&sensorReadings); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "something went wrong", "details": res.Error})
	} else {
		c.JSON(http.StatusOK, ListSensorReadingsResponse{SensorReadings: &sensorReadings, Count: len(sensorReadings)})
	}
}
