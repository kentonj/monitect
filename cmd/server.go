package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/kentonj/monitect/internal/common"
	"github.com/kentonj/monitect/internal/conf"
	"github.com/kentonj/monitect/internal/image"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/sensorreading"
	"github.com/kentonj/monitect/internal/storage"
)

func registerRoutes(
	router *mux.Router,
	sensorClient *sensor.SensorClient,
	sensorReadingClient *sensorreading.SensorReadingClient,
	imageClient *image.ImageClient,
) {
	// ping route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		common.WriteBody(w, http.StatusOK, map[string]string{"msg": "pong"})
	}).Methods("GET")
	// sensor routes
	router.HandleFunc("/sensors", sensorClient.CreateSensor).Methods("POST")
	router.HandleFunc("/sensors/{sensorId}", sensorClient.GetSensor).Methods("GET")
	router.HandleFunc("/sensors/{sensorId}", sensorClient.UpdateSensor).Methods("PUT")
	router.HandleFunc("/sensors/{sensorId}", sensorClient.DeleteSensor).Methods("DELETE")
	router.HandleFunc("/sensors", sensorClient.ListSensors).Methods("GET")
	// sensor-readings routes
	router.HandleFunc("/sensors/{sensorId}/readings", sensorReadingClient.CreateSensorReading).Methods("POST")
	router.HandleFunc("/sensors/{sensorId}/readings", sensorReadingClient.ListSensorReadings).Methods("GET")
	// image routes
	router.HandleFunc("/sensors/{sensorId}/images", imageClient.CreateImage).Methods("POST")
	router.HandleFunc("/sensors/{sensorId}/images/:imageId", imageClient.GetImage).Methods("GET")
	router.HandleFunc("/sensors/{sensorId}/images", imageClient.ListImages).Methods("GET")
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start a timer here
		start := time.Now()
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
		log.Printf("%s %s | %s", r.Method, r.RequestURI, time.Since(start))
	})
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
	confPath := os.Args[1]
	log.Printf("using config path %s", confPath)
	// load config
	config := conf.NewConfig(confPath)
	// open the sqlite database, pass it to the models
	db := storage.Connect(config)
	// assign the db to the models so that it's accessible within the model classes
	if config.Database.Debug {
		db = db.Debug()
	}

	sensorClient := sensor.NewSensorClient(db)
	sensorReadingClient := sensorreading.NewSensorReadingClient(db)
	imageClient := image.NewImageClient(db)

	imageCleaner := NewImageCleaner(sensorClient, imageClient)
	go imageCleaner.Clean(1*time.Hour, 24*time.Hour)

	router := mux.NewRouter()
	registerRoutes(
		router,
		sensorClient,
		sensorReadingClient,
		imageClient,
	)
	router.Use(requestLogger)
	log.Println("registered all routes")
	http.Handle("/", router)
	hostPort := config.Server.Host + ":" + config.Server.Port
	log.Println("serving @", hostPort)
	http.ListenAndServe(hostPort, nil)
}
