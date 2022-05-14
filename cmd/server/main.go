package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kentonj/monitect/internal/conf"
	"github.com/kentonj/monitect/internal/image"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/sensorreading"
	"github.com/kentonj/monitect/internal/storage"
)

func registerRoutes(
	router *gin.Engine,
	sensorClient *sensor.SensorClient,
	sensorReadingClient *sensorreading.SensorReadingClient,
	imageClient *image.ImageClient,
) {
	// ping route
	router.GET("/", func(c *gin.Context) { c.JSON(http.StatusAccepted, gin.H{"msg": "pong"}) })
	// sensor routes
	router.POST("/sensors", sensorClient.CreateSensor)
	router.GET("/sensors/:sensorId", sensorClient.GetSensor)
	router.PUT("/sensors/:sensorId", sensorClient.UpdateSensor)
	router.DELETE("/sensors/:sensorId", sensorClient.DeleteSensor)
	router.GET("/sensors", sensorClient.ListSensors)
	// sensor-readings routes
	router.POST("/sensors/:sensorId/readings", sensorReadingClient.CreateSensorReading)
	router.GET("/sensors/:sensorId/readings", sensorReadingClient.ListSensorReadings)
	// image routes
	router.POST("/sensors/:sensorId/images", imageClient.CreateImage)
	router.GET("/sensors/:sensorId/images/:imageId", imageClient.GetImage)
	router.GET("/sensors/:sensorId/images", imageClient.ListImages)
}

type ImageCleaner struct {
	sensorClient *sensor.SensorClient
	imageClient  *image.ImageClient
}

func NewImageCleaner(sensorClient *sensor.SensorClient, imageClient *image.ImageClient) *ImageCleaner {
	return &ImageCleaner{
		sensorClient: sensorClient,
		imageClient:  imageClient,
	}
}

func (cleaner *ImageCleaner) Clean(interval time.Duration, lookback time.Duration) {
	for {
		time.Sleep(interval)
		oldest := time.Now().Add(-lookback)
		log.Printf("deleting images older than %s", oldest)
		cameras, err := cleaner.sensorClient.ListCameras()
		if err != nil {
			log.Fatal(err)
		}
		for _, cam := range cameras {
			cleaner.imageClient.TruncateImages(cam.ID, oldest)
		}
	}
}

func main() {
	// load config
	config := conf.NewConfig("./configs/server.yaml")
	// open the sqlite database, pass it to the models
	db := storage.Connect(config)
	// assign the db to the models so that it's accessible within the model classes
	if config.Database.Debug {
		db = db.Debug()
	}

	router := gin.Default()
	sensorClient := sensor.NewSensorClient(db)
	sensorReadingClient := sensorreading.NewSensorReadingClient(db)
	imageClient := image.NewImageClient(db)

	imageCleaner := NewImageCleaner(sensorClient, imageClient)
	go imageCleaner.Clean(1*time.Hour, 24*time.Hour)

	registerRoutes(
		router,
		sensorClient,
		sensorReadingClient,
		imageClient,
	)

	router.Run()
}
