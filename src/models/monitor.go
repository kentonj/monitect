package models

import (
	"errors"

	"github.com/google/uuid"
)

type Monitor struct {
	Base
	Name      string  `json:"name" gorm:"unique"`
	Target    float64 `json:"target"`
	Tolerance float64 `json:"tolerance"`
}

func (m *Monitor) Update(body *UpdateMonitorBody) {
	if body.Name != nil {
		m.Name = *body.Name
	}
	if body.Name != nil {
		m.Target = *body.Target
	}
	if body.Name != nil {
		m.Tolerance = *body.Tolerance
	}
}

func CreateMonitor(createMonitorBody *CreateMonitorBody) (*Monitor, error) {
	// create a new monitor oject
	if createMonitorBody.Name == nil || createMonitorBody.Target == nil || createMonitorBody.Tolerance == nil {
		return nil, errors.New("something was null, can't do that homie")
	}
	newMonitor := Monitor{
		Name:      *createMonitorBody.Name,
		Target:    *createMonitorBody.Target,
		Tolerance: *createMonitorBody.Tolerance,
	}
	if res := DB.Create(&newMonitor); res.Error != nil {
		return nil, res.Error
	} else {
		return &newMonitor, nil
	}
}

func UpdateMonitor(id uuid.UUID, updateMonitoryBody *UpdateMonitorBody) error {
	// update a monitor with non-null values provided in the update monitor body
	var monitor Monitor
	if res := DB.First(&monitor, id).Updates(&updateMonitoryBody); res.Error != nil {
		return res.Error
	} else {
		return nil
	}
}

func ListMonitors() (*[]Monitor, error) {
	// return an array of monitors, with nothing in there initially
	monitors := make([]Monitor, 0)
	if res := DB.Find(&monitors); res.Error != nil {
		return nil, res.Error
	} else {
		return &monitors, nil
	}
}
