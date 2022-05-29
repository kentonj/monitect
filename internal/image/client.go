package image

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kentonj/monitect/internal/common"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/storage"
	"gorm.io/gorm"
)

type ImageClient struct {
	db *gorm.DB
}

type Image struct {
	storage.Base
	Bytes    []byte        `json:"-"`
	SensorID uuid.UUID     `json:"-"`
	Sensor   sensor.Sensor `json:"-"`
}

func NewImage(sensorId uuid.UUID, bytes []byte) *Image {
	// create a new image
	img := Image{
		SensorID: sensorId,
		Bytes:    bytes,
	}
	img.AssignUUID()
	return &img
}

func NewImageClient(db *gorm.DB) *ImageClient {
	if err := db.AutoMigrate(&Image{}); err != nil {
		log.Fatal("unable to migrate Image object")
	}
	db.Migrator().CreateConstraint(&Image{}, "Sensor")
	db.Migrator().CreateConstraint(&Image{}, "fk_sensor_readings_sensors")
	// create a new image client
	return &ImageClient{
		db: db,
	}
}

func (client *ImageClient) CreateImage(w http.ResponseWriter, r *http.Request) {
	sensorId, err := uuid.Parse(mux.Vars(r)["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "bad sensorId"})
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"msg": "unable to read form data"})
		return
	}
	defer file.Close()
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"msg": "unable to put file bytes on buffer"})
		return
	}
	image := NewImage(sensorId, buffer.Bytes())
	if res := client.db.Create(&image); res.Error != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "got some conflict", "err": res.Error})
		return
	} else {
		common.WriteBody(w, http.StatusOK, common.AnyMap{"imageId": image.ID})
		return
	}
}

func (client *ImageClient) GetImage(w http.ResponseWriter, r *http.Request) {
	// Get a specified image, return just the bytes
	vars := mux.Vars(r)
	sensorId, err := uuid.Parse(vars["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "bad uuid"})
		return
	}
	imageIdString := vars["imageId"]
	var query *gorm.DB
	if imageIdString == "latest" {
		query = client.db.Where("sensor_id = ?", sensorId).Order("created_at desc")
	} else {
		imageId, err := uuid.Parse(imageIdString)
		if err != nil {
			common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "bad uuid"})
			return
		}
		query = client.db.Where("sensor_id = ?", sensorId).Where("image_id = ?", imageId)
	}
	var image Image
	if res := query.First(&image); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"msg": "something happened", "err": res.Error})
		return
	} else {
		common.WriteData(w, http.StatusOK, "image/png", image.Bytes)
		return
	}
}

type ListImagesResponse struct {
	Images *[]Image
	Count  int
}

func (client *ImageClient) ListImages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sensorId, err := uuid.Parse(vars["sensorId"])
	if err != nil {
		common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": "bad uuid"})
		return
	}
	limitString := vars["limit"]
	var limit int
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			common.WriteBody(w, http.StatusBadRequest, common.AnyMap{"msg": fmt.Sprint("invalid limit ", limitString)})
			return
		}
	} else {
		limit = 100
	}
	query := client.db.Where("sensor_id = ?", sensorId).Order("created_at desc").Limit(limit)
	images := make([]Image, 0)
	if res := query.Find(&images); res.Error != nil {
		common.WriteBody(w, http.StatusInternalServerError, common.AnyMap{"msg": "something went wrong", "details": res.Error})
		return
	} else {
		common.WriteBody(w, http.StatusOK, ListImagesResponse{Images: &images, Count: len(images)})
		return
	}
}

// clean up images that are older than the specified date for all sensors of type 'camera'
func (client *ImageClient) TruncateImages(sensorId uuid.UUID, oldest time.Time) error {
	query := client.db.Unscoped().Where("sensor_id = ?", sensorId).Where("created_at < ?", oldest)
	if res := query.Delete(&[]Image{}); res.Error != nil {
		log.Printf("deleted %d rows", res.RowsAffected)
		return res.Error
	} else {
		return nil
	}
}
