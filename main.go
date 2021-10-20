package main

import (
	"monitect/conf"
	"monitect/models"
	"monitect/services"
	"monitect/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// load config
	config := conf.NewConfig("./config.yaml")
	// open the mongo connection, we'll pass the mongo connection to the necessary handlers
	db := storage.Connect(config)
	defer db.Close()
	models.DB = db
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.POST("/sensors", services.CreateSensor)
	router.Run()
}
