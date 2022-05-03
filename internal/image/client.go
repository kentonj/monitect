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
	Bytes    []byte        `json:"-"`
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
	// Get a specified image, return just the bytes
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
		c.Data(http.StatusOK, "image/png", image.Bytes)
		return
	}
}

type ListImagesResponse struct {
	Images *[]Image
	Count  int
}

func (client *ImageClient) ListImages(c *gin.Context) {
	sensorId, err := uuid.Parse(c.Param("sensorId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad uuid"})
		return
	}
	query := client.db.Where("sensor_id = ?", sensorId).Order("created_at desc")
	images := make([]Image, 0)
	if res := query.Find(&images); res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "something went wrong", "details": res.Error})
		return
	} else {
		c.JSON(http.StatusOK, ListImagesResponse{Images: &images, Count: len(images)})
		return
	}
}

// clean up images that are older than the specified date for all sensors of type 'camera'
func (client *ImageClient) TruncateImages(sensorId uuid.UUID, oldest time.Time) error {
	query := client.db.Unscoped().Where("sensor_id = ?", sensorId).Where("created_at < ?", oldest)
	if res := query.Delete(&[]Image{}); res.Error != nil {
		log.Printf("deleted %d rows", res.RowsAffected)
		return res.Error
	} else {
		return nil
	}
}
