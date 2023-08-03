package sensor

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	storage "github.com/kentonj/monitect/internal/storage"
	"github.com/kentonj/monitect/internal/stream"
	"gorm.io/gorm"
)

type SensorClient struct {
	db      *gorm.DB
	sensors map[string]*Sensor // sensors are always cached. When a client requests a sensorId, it should be retrieved from this cache
}

type SensorType string

type SensorReadingType string

const (
	sensorReadingMonitorName = "_SENSOR_READING_MONITOR"
)

var (
	// sensor types
	ThermometerType = SensorType("thermometer")
	CameraType      = SensorType("camera")
	GenericType     = SensorType("generic") // generic type
	// reading types
	Base64Type    = SensorReadingType("base64")
	FloatType     = SensorReadingType("float64")
	InterfaceType = SensorReadingType("interface")
)

type Sensor struct {
	storage.Base
	Name          string            `json:"name,omitempty"`
	Type          SensorType        `json:"type,omitempty"`
	ReadingType   SensorReadingType `json:"readingType,omitempty"`
	Unit          string            `json:"unit,omitempty"`
	StreamManager *stream.Manager   `json:"-" gorm:"-"` // don't include the stream in database operations or json
}

type SensorReading struct {
	storage.Base
	Sensor   Sensor    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	SensorID uuid.UUID `json:"sensorId"`
	Value    []byte    `json:"value"`
}

type CreateSensorBody struct {
	Name        string            `json:"name"`
	Type        SensorType        `json:"type"`
	ReadingType SensorReadingType `json:"readingType"`
}

type UpdateSensorBody struct {
	Name        string            `json:"name"`
	Type        SensorType        `json:"type"`
	ReadingType SensorReadingType `json:"readingType"`
}

type ListSensorsResponse struct {
	Msg     string   `json:"msg"`
	Sensors []Sensor `json:"sensors"`
	Count   int      `json:"count"`
}

func NewSensorClient(db *gorm.DB) *SensorClient {
	if err := db.AutoMigrate(&Sensor{}, &SensorReading{}); err != nil {
		log.Fatal("Could not automigrate the sensor object")
	} else {
		log.Println("Migrated Sensor and SensorReading")
	}
	client := SensorClient{db: db, sensors: make(map[string]*Sensor)}
	if sensors, err := client.listSensorsFromDb(); err != nil {
		log.Fatalf("unable to list sensors: %s", err)
	} else {
		// add the stream manager for each sensor
		for _, sensor := range sensors {
			sensor.StreamManager = stream.NewManager(100)
			client.addSensorToCache(sensor)
			go client.monitorReadings(sensor, 10*time.Second)
		}
	}
	return &client
}

// parseSensorReadingValue reads the byte array into the correct data type based on the provided SensorReadingType
func parseSensorReadingValue(value []byte, readingType SensorReadingType) (interface{}, error) {
	switch readingType {
	case Base64Type:
		var s string
		if err := json.Unmarshal(value, &s); err != nil {
			return nil, err
		} else {
			return s, nil
		}
	case FloatType:
		var f float64
		if err := json.Unmarshal(value, &f); err != nil {
			return nil, err
		} else {
			return f, nil
		}
	default:
		var i interface{}
		if err := json.Unmarshal(value, &i); err != nil {
			return nil, err
		} else {
			return i, nil
		}
	}
}

func NewSensorReading(sensorID uuid.UUID, data []byte) *SensorReading {
	sr := SensorReading{SensorID: sensorID, Value: data}
	sr.AssignUUID()
	return &sr
}

func (s *Sensor) readingType() SensorReadingType {
	switch s.Type {
	case ThermometerType:
		return FloatType
	case CameraType:
		return Base64Type
	default:
		// we don't know the type, so just call it an interface
		return InterfaceType
	}
}

func (s *SensorClient) addSensorToCache(sensor *Sensor) {
	s.sensors[sensor.ID.String()] = sensor
}

// monitorReadings periodically writes the reading to the database
func (s *SensorClient) monitorReadings(sensor *Sensor, interval time.Duration) {
	if err := sensor.StreamManager.AddClient(sensorReadingMonitorName); err != nil {
		log.Fatalf("unable to add sensor reading monitor %s", err)
		return
	}
	monitorStream, _ := sensor.StreamManager.ClientStream(sensorReadingMonitorName)
	dataCh := monitorStream.C()
	ticker := time.NewTicker(interval)
	for range ticker.C {
		select {
		case message := <-dataCh:
			// see if there is more to consume on the ch
			remainingMessages := len(dataCh)
			data := message
			for i := 0; i < remainingMessages; i++ {
				data = <-dataCh
			}
			log.Printf("got latest reading (size %d) on the %s[%s] stream", len(data), sensor.ID, sensorReadingMonitorName)
			newReading := NewSensorReading(sensor.ID, data)
			log.Print("here is the new reading")
			// log.Print(newReading.SensorID)
			// log.Print(newReading.Sensor)
			// log.Print(newReading.Type)
			// overwrite the existing reading, or create a new reading
			var existingReading SensorReading
			if res := s.db.Order("created_at desc").Take(&existingReading); res.Error != nil {
				if res := s.db.Create(newReading); res.Error != nil {
					log.Fatalf("unable to create a new reading: %s", res.Error)
				}
				log.Printf("no previous reading existed")
			} else {
				existingReading.Value = newReading.Value
				if res := s.db.Save(&existingReading); res.Error != nil {
					log.Fatalf("unable to update the sensor reading reading: %s", res.Error)
				}
				log.Printf("updated the last reading")
			}
		default:
			// do nothing
			log.Printf("no new readings on the %s[%s] stream", sensor.ID, sensorReadingMonitorName)
		}
	}
}

func (s *SensorClient) GetLatestReading(sensorId uuid.UUID) (*SensorReading, error) {
	latestReading := SensorReading{SensorID: sensorId}
	if res := s.db.Order("updated_at desc").First(&latestReading); res.Error != nil {
		return nil, res.Error
	} else {
		return &latestReading, nil
	}
}

func (s *Sensor) update(u *UpdateSensorBody) {
	if u.Name != "" {
		s.Name = u.Name
	}
	if u.Type != "" {
		s.Type = u.Type
	}
	if u.ReadingType != "" {
		s.ReadingType = u.ReadingType
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
		Name:        body.Name,
		Type:        body.Type,
		ReadingType: body.ReadingType,
	}
	sensor.AssignUUID()
	return &sensor, nil
}

// CreateSensor creates a Sensor from the CreateSensorBody and adds it to the sensor cache and database
func (s *SensorClient) CreateSensor(createSensorBody *CreateSensorBody) (*Sensor, error) {
	sensor, err := createSensorBody.toSensor()
	if err != nil {
		return nil, err
	}
	if res := s.db.Create(sensor); res.Error != nil {
		return nil, res.Error
	}
	// assign a stream.Manager for the new sensor
	sensor.StreamManager = stream.NewManager(1000)
	s.sensors[sensor.ID.String()] = sensor
	go s.monitorReadings(sensor, 10*time.Second)
	return sensor, nil
}

// GetSensor retrieves the sensor out of the sensor cache
func (s *SensorClient) GetSensor(sensorId string) (*Sensor, bool) {
	sensor, found := s.sensors[sensorId]
	return sensor, found
}

// PublishSensorReading sends a message to the sensor's streamManager
func (s *SensorClient) PublishSensorReading(sensorId string, data []byte) error {
	sensor, found := s.GetSensor(sensorId)
	if !found {
		return fmt.Errorf("sensor %s not found", sensorId)
	}
	sensor.StreamManager.Send(data)
	return nil
}

// Add a client to the sensor, and return the client stream
func (s *SensorClient) AddClientStream(sensorId string, clientId string) (*stream.Stream, error) {
	sensor, found := s.GetSensor(sensorId)
	if !found {
		return nil, fmt.Errorf("sensor %s not found", sensorId)
	}
	if err := sensor.StreamManager.AddClient(clientId); err != nil {
		return nil, err
	}
	if clientStream, found := sensor.StreamManager.ClientStream(clientId); !found {
		panic(fmt.Errorf("client stream for sensor not found"))
	} else {
		return clientStream, nil
	}
}

func (s *SensorClient) RemoveClientStream(sensorId string, clientId string) {
	sensor, found := s.GetSensor(sensorId)
	if !found {
		log.Println("tried to remove a client stream from a sensor that didn't exist")
		return
	}
	sensor.StreamManager.RemoveClient(clientId)
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
	delete(s.sensors, sensor.ID.String())
	return nil
}

func (s *SensorClient) UpdateSensor(sensorId string, updateSensorBody *UpdateSensorBody) (*Sensor, error) {
	sensor := s.sensors[sensorId]
	sensor.update(updateSensorBody)
	if res := s.db.Save(&sensor); res.Error != nil {
		return nil, res.Error
	} else {
		return sensor, nil
	}
}

// listSensorsFromDb returns the sensors from the database
func (s *SensorClient) listSensorsFromDb() ([]*Sensor, error) {
	var sensors []*Sensor
	if res := s.db.Find(&sensors); res.Error != nil {
		return nil, res.Error
	} else {
		return sensors, nil
	}
}

// ListSensors returns the sensors from the cache
func (s *SensorClient) ListSensors() []*Sensor {
	sensors := make([]*Sensor, 0)
	for _, sensor := range s.sensors {
		sensors = append(sensors, sensor)
	}
	return sensors
}
