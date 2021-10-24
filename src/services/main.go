package services

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// sensor routes
	router.POST("/sensors", CreateSensor)
	router.GET("/sensors", ListSensors)
	router.GET("/sensors/:sensorId", GetSensor)
	// sensor-readings routes
	router.POST("/sensors/:sensorId/readings", CreateSensorReading)
	router.GET("/sensors/:sensorId/readings", ListSensorReadings)
}
