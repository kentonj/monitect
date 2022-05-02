package sensor

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	storage "github.com/kentonj/monitect/internal/storage"
)

// create a new sensor client
// this should also automigrate the db object, and set up any necessary foreign keys
// this should then be returned to any of the HTTP handlers, or RPC handlers, or anything like that
type SensorClient struct {
	db *gorm.DB
}

func NewSensorClient(db *gorm.DB) *SensorClient {
	if err := db.AutoMigrate(&Sensor{}); err != nil {
		log.Fatal("Could not automigrate the sensor object")
	}
	client := SensorClient{db: db}
	return &client
}

type Sensor struct {
	storage.Base
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

func (s *Sensor) Update(u *UpdateSensorBody) {
	if u.Name != "" {
		s.Name = u.Name
	}
	if u.Type != "" {
		s.Type = u.Type
	}
}

type CreateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// convert the sensor body to a sensor, doing any necessary logical checks
func (body *CreateSensorBody) toSensor() (*Sensor, error) {
	if body.Name == "" {
		return nil, errors.New("name cannot be nil")
	}
	if body.Type == "" {
		return nil, errors.New("type cannot be nil")
	}
	sensor := Sensor{
		Name: body.Name,
		Type: body.Type,
	}
	sensor.AssignUUID()
	return &sensor, nil
}

// create a sensor from a createsensor body
func (client *SensorClient) CreateSensor(c *gin.Context) {
	var createSensorBody CreateSensorBody
	if parseErr := c.ShouldBindJSON(&createSensorBody); parseErr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	sensor, err := createSensorBody.toSensor()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err})
		return
	}
	if res := client.db.Create(sensor); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": res.Error})
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{"sensor": sensor})
		return
	}
}

// get a sensor by it's ID
func (client *SensorClient) GetSensor(c *gin.Context) {
	id, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "not a valid uuid"})
		return
	}
	var sensor Sensor
	if res := client.db.First(&sensor, id); res.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "OK", "sensor": sensor})
		return
	}
}

type UpdateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (client *SensorClient) UpdateSensor(c *gin.Context) {
	id, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "not a valid uuid"})
		return
	}
	var updateSensorBody UpdateSensorBody
	if err := c.ShouldBindJSON(&updateSensorBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err})
		return
	}
	var sensor Sensor
	if res := client.db.First(&sensor, id); res.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	sensor.Update(&updateSensorBody)
	if res := client.db.Save(&sensor); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": res.Error})
	}
}

type ListSensorsResponse struct {
	Msg     string   `json:"msg"`
	Sensors []Sensor `json:"sensors"`
	Count   int      `json:"count"`
}

// list sensors
func (client *SensorClient) ListSensors(c *gin.Context) {
	sensors := make([]Sensor, 0)
	if res := client.db.Find(&sensors); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": res.Error})
		return
	} else {
		c.JSON(http.StatusOK, ListSensorsResponse{
			Msg:     "OK",
			Sensors: sensors,
			Count:   len(sensors),
		})
	}
}
