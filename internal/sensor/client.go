package sensor

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	storage "github.com/kentonj/monitect/internal/storage"
	"github.com/kentonj/monitect/internal/stream"
	"gorm.io/gorm"
)

type SensorClient struct {
	db            *gorm.DB
	StreamManager *stream.Manager
}

type Sensor struct {
	storage.Base
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Unit string `json:"unit,omitempty"`
}

type CreateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Unit string `json:"unit"`
}

type UpdateSensorBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Unit string `json:"unit"`
}

type ListSensorsResponse struct {
	Msg     string   `json:"msg"`
	Sensors []Sensor `json:"sensors"`
	Count   int      `json:"count"`
}

func NewSensorClient(db *gorm.DB) *SensorClient {
	if err := db.AutoMigrate(&Sensor{}); err != nil {
		log.Fatal("Could not automigrate the sensor object")
	} else {
		log.Println("Migrated sensors")
	}
	streamManager := stream.NewManager(1000)
	client := SensorClient{db: db, StreamManager: streamManager}
	if sensors, err := client.ListSensors(); err != nil {
		log.Fatalf("unable to list sensors: %s", err)
	} else {
		// add sensors to the stream manager
		for _, sensor := range sensors {
			log.Printf("adding sensor %s to stream manager", sensor.ID)
			if err := client.StreamManager.Add(sensor.ID.String()); err != nil {
				log.Fatalf("unable to add sensor %s to streamManager: %s", sensor.ID.String(), err)
			}
		}
	}
	return &client
}

func (s *Sensor) update(u *UpdateSensorBody) {
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

// CreateSensor creates a Sensor from the CreateSensorBody
func (s *SensorClient) CreateSensor(createSensorBody *CreateSensorBody) (*Sensor, error) {
	sensor, err := createSensorBody.toSensor()
	if err != nil {
		return nil, err
	}
	if res := s.db.Create(sensor); res.Error != nil {
		return nil, res.Error
	}
	if err := s.StreamManager.Add(sensor.ID.String()); err != nil {
		return nil, err
	}
	return sensor, nil
}

// GetSensor retrieves a sensor by it's ID
func (s *SensorClient) GetSensor(sensorId string) (*Sensor, error) {
	sensorUUID, err := uuid.Parse(sensorId)
	if err != nil {
		return nil, err
	}
	var sensor Sensor
	if res := s.db.First(&sensor, sensorUUID); res.Error != nil {
		return nil, res.Error
	} else {
		return &sensor, nil
	}
}

// GetSensorStream gets the stream for the sensor
func (s *SensorClient) GetSensorStream(sensorId string) (*stream.Stream, bool) {
	return s.StreamManager.Stream(sensorId)
}

func (s *SensorClient) PublishSensorReading(sensorId string, data []byte) error {
	sensorStream, found := s.StreamManager.Stream(sensorId)
	if !found {
		return fmt.Errorf("sensor stream for %s not found", sensorId)
	}
	sensorStream.Publish(data)
	return nil
}

func (s *SensorClient) AddClientStream(sensorId string, clientId string) (chan []byte, error) {
	sensorStream, found := s.StreamManager.Stream(sensorId)
	if !found {
		return nil, fmt.Errorf("stream for sensor not found")
	}
	if err := sensorStream.AddClient(clientId); err != nil {
		return nil, err
	}
	if clientStream, found := sensorStream.Client(clientId); !found {
		return nil, fmt.Errorf("client stream for sensor not found")
	} else {
		return clientStream, nil
	}
}

func (s *SensorClient) RemoveClientStream(sensorId string, clientId string) {
	log.Println("starting to remove client stream")
	sensorStream, found := s.StreamManager.Stream(sensorId)
	if !found {
		log.Println("tried to remove a client stream that didn't exist")
		return
	}
	sensorStream.RemoveClient(clientId)
}

// DeleteSensor deletes a sensor by its id
func (s *SensorClient) DeleteSensor(sensorId string) error {
	sensorUUID, err := uuid.Parse(sensorId)
	if err != nil {
		return err
	}
	var sensor Sensor
	if res := s.db.Delete(&sensor, sensorUUID); res.Error != nil {
		return res.Error
	}
	s.StreamManager.Remove(sensorId)
	return nil
}

func (s *SensorClient) UpdateSensor(sensorId string, updateSensorBody *UpdateSensorBody) (*Sensor, error) {
	sensorUUID, err := uuid.Parse(sensorId)
	if err != nil {
		return nil, err
	}
	var sensor Sensor
	if res := s.db.First(&sensor, sensorUUID); res.Error != nil {
		return nil, res.Error
	}
	sensor.update(updateSensorBody)
	if res := s.db.Save(&sensor); res.Error != nil {
		return nil, res.Error
	} else {
		return &sensor, nil
	}
}

// list sensors
func (s *SensorClient) ListSensors() ([]*Sensor, error) {
	sensors := make([]*Sensor, 0)
	if res := s.db.Find(&sensors); res.Error != nil {
		return nil, res.Error
	} else {
		return sensors, nil
	}
}
