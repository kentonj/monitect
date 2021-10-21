package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kentonj/monitect/src/conf"
	"github.com/kentonj/monitect/src/models"
	"github.com/kentonj/monitect/src/services"
	"github.com/kentonj/monitect/src/storage"
)

func main() {
	// load config
	config := conf.NewConfig("./conf/server.yaml")
	// open the mongo connection, we'll pass the mongo connection to the necessary handlers
	conn := storage.Connect(config)
	defer conn.Close()
	// assign the db to the models so that it's accessible within the model classes
	models.DB = conn.Db()
	// initialize router and register routes
	router := gin.Default()
	services.RegisterRoutes(router)
	router.Run()
}
