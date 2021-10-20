package main

import (
	"fmt"
	"growmon/conf"
	"growmon/models"
	"growmon/storage"
)

func main() {
	// load config
	config := conf.NewConfig("./config.yaml")
	// open the mongo connection, we'll pass the mongo connection to the necessary handlers
	models.DB = storage.Connect(config)
	s := models.CreateSensor("anotherone", "temperature")
	retrievedSensor := models.GetSensorByName(s.Name)
	fmt.Println(s)
	fmt.Println(retrievedSensor)
}
