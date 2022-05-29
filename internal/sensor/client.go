package sensor

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/kentonj/monitect/internal/common"
	storage "github.com/kentonj/monitect/internal/storage"
)

// create a new sensor client
// this should also automigrate the db object, and set up any necessary foreign keys
// this should then be returned to any of the HTTP handlers, or RPC handlers, or anything like that
type SensorClient struct {
	db *gorm.DB
}

func NewSensorClient(db *gorm.DB) *SensorClient {
	if err := db.AutoMigrate(&Sensor{}); err != nil {
		log.Fatal("Could not automigrate the sensor object")
	} else {
		log.Println("Migrated sensors")
	}
	client := SensorClient{db: db}
	return &client
}

type Sensor struct {
	storage.Base
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Unit string `json:"unit,omitempty"`
}

func (s *Sensor) Update(u *UpdateSensorBody) {
	if u.Name != "" {
		s.Name = u.Name
	}
	if u.Type != "" {
		s.Type = u.Type
	}
	if u.Unit != "" {
		s.Unit = u.Unit
	}
}

type CreateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Unit string `json:"unit"`
}

// convert the sensor body to a sensor, doing any necessary logical checks
func (body *CreateSensorBody) toSensor() (*Sensor, error) {
	if body.Name == "" {
		return nil, errors.New("name cannot be nil")
	}
	if body.Type == "" {
		return nil, errors.New("type cannot be nil")
	}
	sensor := Sensor{
		Name: body.Name,
		Type: body.Type,
		Unit: body.Unit,
	}
	sensor.AssignUUID()
	return &sensor, nil
}

// create a sensor from a createsensor body
func (client *SensorClient) CreateSensor(w http.ResponseWriter, r *http.Request) {
	var createSensorBody CreateSensorBody
	if parseErr := common.BindJSON(r, &createSensorBody); parseErr != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": parseErr.Error()})
		return
	}
	sensor, err := createSensorBody.toSensor()
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": err.Error()})
		return
	}
	if res := client.db.Create(sensor); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"err": res.Error.Error()})
		return
	} else {
		common.WriteBody(w, http.StatusCreated, common.AnyMap{"sensor": sensor})
		return
	}
}

// get a sensor by it's ID
func (client *SensorClient) GetSensor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": "not a valid uuid"})
		return
	}
	var sensor Sensor
	if res := client.db.First(&sensor, id); res.Error != nil {
		common.WriteBody(w, http.StatusNotFound, nil)
		return
	} else {
		common.WriteBody(w, http.StatusOK, common.AnyMap{"msg": "OK", "sensor": sensor})
		return
	}
}

// delete a sensor by it's id
func (client *SensorClient) DeleteSensor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": "not a valid uuid"})
		return
	}
	var sensor Sensor
	if res := client.db.Delete(&sensor, id); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"err": res.Error})
		return
	} else {
		common.WriteBody(w, http.StatusOK, common.AnyMap{"msg": "OK", "sensor": sensor})
		return
	}
}

type UpdateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Unit string `json:"unit"`
}

func (client *SensorClient) UpdateSensor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": "not a valid uuid"})
		return
	}
	var updateSensorBody UpdateSensorBody
	if err := common.BindJSON(r, &updateSensorBody); err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"err": err})
		return
	}
	var sensor Sensor
	if res := client.db.First(&sensor, id); res.Error != nil {
		common.WriteBody(w, http.StatusNotFound, nil)
		return
	}
	sensor.Update(&updateSensorBody)
	if res := client.db.Save(&sensor); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"err": res.Error})
		return
	}
}

type ListSensorsResponse struct {
	Msg     string   `json:"msg"`
	Sensors []Sensor `json:"sensors"`
	Count   int      `json:"count"`
}

// list cameras
func (client *SensorClient) ListCameras() ([]Sensor, error) {
	cameras := make([]Sensor, 0)
	if res := client.db.Where("type = ?", "camera").Find(&cameras); res.Error != nil {
		return nil, res.Error
	} else {
		return cameras, nil
	}
}

// list sensors
func (client *SensorClient) ListSensors(w http.ResponseWriter, r *http.Request) {
	sensors := make([]Sensor, 0)
	if res := client.db.Find(&sensors); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"err": res.Error})
		return
	} else {
		common.WriteBody(w, http.StatusOK, ListSensorsResponse{
			Msg:     "OK",
			Sensors: sensors,
			Count:   len(sensors),
		})
	}
}
