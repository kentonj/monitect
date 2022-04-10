package models

type BaseResponse struct {
	Msg string `json:"msg"`
}

var SUCCESS = BaseResponse{Msg: "OK"}

type CreateSensorBody struct {
	Name *string `json:"name"`
	Type *string `json:"type"`
}

type UpdateSensorBody struct {
	CreateSensorBody
}

type CreateSensorResponse struct {
	BaseResponse
	Sensor Sensor `json:"sensor"`
}

type GetSensorResponse struct {
	Msg    string `json:"msg"`
	Sensor Sensor `json:"sensor"`
}

type ListSensorsResponse struct {
	Msg     string   `json:"msg"`
	Sensors []Sensor `json:"sensors"`
	Count   int      `json:"count"`
}

type CreateSensorReadingBody struct {
	Value *float64 `json:"value"`
}

type CreateSensorReadingResponse struct {
	Msg           string        `json:"msg"`
	SensorReading SensorReading `json:"sensorReading"`
}

type ListSensorReadingsResponse struct {
	Msg            string          `json:"msg"`
	SensorReadings []SensorReading `json:"sensorsReadings"`
	Count          int             `json:"count"`
}

type CreateMonitorBody struct {
	Name      *string
	Target    *float64
	Tolerance *float64
}

type UpdateMonitorBody struct {
	CreateMonitorBody
}
