package image

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/storage"
	"gorm.io/gorm"
)

type ImageClient struct {
	db *gorm.DB
}

type Image struct {
	storage.Base
	Bytes    []byte        `json:"bytes"`
	SensorID uuid.UUID     `json:"-"`
	Sensor   sensor.Sensor `json:"-"`
}

func NewImage(sensorId uuid.UUID, bytes []byte) *Image {
	// create a new image
	img := Image{
		SensorID: sensorId,
		Bytes:    bytes,
	}
	img.AssignUUID()
	return &img
}

func NewImageClient(db *gorm.DB) *ImageClient {
	if err := db.AutoMigrate(&Image{}); err != nil {
		log.Fatal("unable to migrate Image object")
	}
	db.Migrator().CreateConstraint(&Image{}, "Sensor")
	db.Migrator().CreateConstraint(&Image{}, "fk_sensor_readings_sensors")
	// create a new image client
	return &ImageClient{
		db: db,
	}
}

func (client *ImageClient) CreateImage(c *gin.Context) {
	sensorId, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad sensorId"})
		return
	}
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "unable to read form data"})
		return
	}
	defer file.Close()
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "unable to put file bytes on buffer"})
		return
	}
	image := NewImage(sensorId, buffer.Bytes())
	if res := client.db.Create(&image); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "got some conflict", "err": res.Error})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"imageId": image.ID})
		return
	}
}

func (client *ImageClient) GetImage(c *gin.Context) {
	// Get a specified image
	sensorId, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad uuid"})
		return
	}
	imageIdString := c.Param("imageId")
	var query *gorm.DB
	if imageIdString == "latest" {
		query = client.db.Where("sensor_id = ?", sensorId).Order("created_at desc")
	} else {
		imageId, err := uuid.Parse(imageIdString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad uuid"})
			return
		}
		query = client.db.Where("sensor_id = ?", sensorId).Where("image_id = ?", imageId)
	}
	var image Image
	if res := query.First(&image); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "something happened", "err": res.Error})
		return
	} else {
		c.JSON(http.StatusOK, &image)
		return
	}
}

// clean up images that are older than the specified date
func (client *ImageClient) TruncateImages(c *gin.Context) {
	sensorId, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad uuid"})
		return
	}
	oldest := c.Query("oldest")
	if oldest == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "you must specify an `oldest` date"})
		return
	}
	t, err := time.Parse(time.RFC3339, oldest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "RFC3339 time format required"})
		return
	}
	query := client.db.Where("sensor_id = ?", sensorId).Where("created_at < ?", t)
	if res := query.Delete(&[]Image{}); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "something went wrong", "err": res.Error})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "nice"})
		return
	}
}
