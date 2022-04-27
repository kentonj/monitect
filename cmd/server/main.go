package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kentonj/monitect/internal/conf"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/sensorreading"
	"github.com/kentonj/monitect/internal/storage"
)

func registerRoutes(
	router *gin.Engine,
	sensorClient *sensor.SensorClient,
	sensorReadingClient *sensorreading.SensorReadingClient,
) {
	// sensor routes
	router.POST("/sensors", sensorClient.CreateSensor)
	router.GET("/sensors/:sensorId", sensorClient.GetSensor)
	router.PUT("/sensors/:sensorId", sensorClient.UpdateSensor)
	router.GET("/sensors", sensorClient.ListSensors)
	// // sensor-readings routes
	router.POST("/sensors/:sensorId/readings", sensorReadingClient.CreateSensorReading)
	router.GET("/sensors/:sensorId/readings", sensorReadingClient.ListSensorReadings)
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

	registerRoutes(
		router,
		sensorClient,
		sensorReadingClient,
	)
	router.Run()
}