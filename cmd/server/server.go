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
	// open the sqlite database, pass it to the models
	db := storage.Connect(config)
	// assign the db to the models so that it's accessible within the model classes
	if config.Database.Debug {
		db = db.Debug()
	}
	models.DB = db
	// make sure that schemas are up-to-date
	models.MigrateSchemas()
	// initialize router and register routes
	router := gin.Default()
	services.RegisterRoutes(router)
	router.Run()
}
