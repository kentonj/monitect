package sensorreading

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kentonj/monitect/internal/common"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/storage"
	"gorm.io/gorm"
)

type SensorReading struct {
	storage.Base
	Value    *float64      `json:"value"`
	SensorID uuid.UUID     `json:"-"`
	Sensor   sensor.Sensor `json:"-"`
}

type SensorReadingClient struct {
	db *gorm.DB
}

// return a new sensor reading client
func NewSensorReadingClient(db *gorm.DB) *SensorReadingClient {
	if err := db.AutoMigrate(&SensorReading{}); err != nil {
		log.Fatal("unable to migrate SensorReading object")
	}
	db.Migrator().CreateConstraint(&SensorReading{}, "Sensor")
	db.Migrator().CreateConstraint(&SensorReading{}, "fk_sensor_readings_sensors")
	return &SensorReadingClient{
		db: db,
	}
}

type CreateSensorReadingResponse struct {
	SensorReading *SensorReading `json:"sensorReading"`
}

func (client *SensorReadingClient) CreateSensorReading(w http.ResponseWriter, r *http.Request) {
	sensorId, err := uuid.Parse(mux.Vars(r)["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": "unable to parse uuid"})
		return
	}
	var sensorReading SensorReading
	if err := common.BindJSON(r, &sensorReading); err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": "bad json input"})
		return
	}
	if sensorReading.Value == nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "sensor reading value cannot be nil"})
		return
	}
	sensorReading.SensorID = sensorId
	sensorReading.AssignUUID()
	if res := client.db.Create(&sensorReading); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"msg": "whoops, something went wrong", "details": res.Error})
		return
	} else {
		common.WriteBody(w, http.StatusOK, CreateSensorReadingResponse{SensorReading: &sensorReading})
		return
	}
}

type ListSensorReadingsResponse struct {
	SensorReadings *[]SensorReading `json:"sensorReadings"`
	Count          int              `json:"count"`
}

func (client *SensorReadingClient) ListSensorReadings(w http.ResponseWriter, r *http.Request) {
	sensorId, err := uuid.Parse(mux.Vars(r)["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "bad uuid"})
		return
	}
	limitString := r.URL.Query().Get("limit")
	var limit int
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": fmt.Sprint("invalid limit ", limitString)})
			return
		}
	} else {
		limit = 1000
	}
	query := client.db.Where("sensor_id = ?", sensorId).Order("created_at desc").Limit(limit)
	sensorReadings := make([]SensorReading, 0)
	if res := query.Find(&sensorReadings); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"msg": "something went wrong", "details": res.Error})
		return
	} else {
		common.WriteBody(w, http.StatusOK, ListSensorReadingsResponse{SensorReadings: &sensorReadings, Count: len(sensorReadings)})
		return
	}
}
