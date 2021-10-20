package services

import (
	"fmt"
	"monitect/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CreateSensorResponse struct {
	Msg string `json:"msg"`
}

func CreateSensor(c *gin.Context) {
	// register a sensor, error if it already exists
	newSensorBody := new(CreateSensorBody)
	if err := c.ShouldBindJSON(&newSensorBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existingSensor := models.GetSensorByName(newSensorBody.Name)
	if existingSensor != nil {
		// sensor already exists, so return an appropriate error code
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("sensor with name %s already exists", newSensorBody.Name)})
		return
	}
	models.CreateSensor(newSensorBody.Name, newSensorBody.Type)
	c.JSON(http.StatusCreated, CreateSensorResponse{Msg: "Successfully created sensor"})
}
