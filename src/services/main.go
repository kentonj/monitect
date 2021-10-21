package services

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

func ParamsToFilter(params map[string][]string) bson.D {
	// take query params and turn it into a filter
	filter := bson.D{}
	for name, values := range params {
		for _, value := range values {
			fmt.Println("adding a value to map")
			addition := bson.E{Key: name, Value: value}
			fmt.Println(addition)
			filter = append(filter, addition)
		}
	}
	return filter
}
