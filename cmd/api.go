package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kentonj/monitect/internal/common"
	"github.com/kentonj/monitect/internal/conf"
	"github.com/kentonj/monitect/internal/sensor"
	"github.com/kentonj/monitect/internal/storage"
	"github.com/kentonj/monitect/internal/stream"
	"gorm.io/gorm"
)

type Monitect struct {
	upgrader     websocket.Upgrader
	router       *mux.Router
	streamRPS    time.Duration
	SensorClient *sensor.SensorClient
}

type SensorReadingBody struct {
	Value float64 `json:"value"`
}

type ErrorResponse struct {
	Msg string `json:"msg"`
	Err error  `json:"err"`
}

type ListSensorsResponse struct {
	Sensors []*sensor.Sensor `json:"sensors"`
	Count   int              `json:"count"`
}

func (m *Monitect) registerRoutes() {
	// ping route
	m.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		common.WriteBody(w, http.StatusOK, map[string]string{"msg": "pong"})
	}).Methods("GET")
	// sensor routes
	m.router.HandleFunc("/sensors", m.createSensor).Methods("POST")
	m.router.HandleFunc("/sensors", m.listSensors).Methods("GET")
	m.router.HandleFunc("/sensors/{sensorId}", m.getSensor).Methods("GET")
	m.router.HandleFunc("/sensors/{sensorId}/reading", m.publishSensorReading).Methods("POST")
	m.router.HandleFunc("/sensors/{sensorId}", m.updateSensor).Methods("PUT")
	m.router.HandleFunc("/sensors/{sensorId}", m.deleteSensor).Methods("DELETE")
	// websocket routes
	m.router.HandleFunc("/sensors/{sensorId}/read", m.readSensorFeed).Methods("GET")
	m.router.HandleFunc("/sensors/{sensorId}/publish", m.publishSensorFeed).Methods("GET")
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
		log.Printf("%s %s | %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func NewMonitect(db *gorm.DB) *Monitect {
	router := mux.NewRouter()
	sensorClient := sensor.NewSensorClient(db)
	// list all sensors, and add them to the stream client
	streamRPS := time.Second / 30
	return &Monitect{
		upgrader:     websocket.Upgrader{},
		router:       router,
		streamRPS:    streamRPS,
		SensorClient: sensorClient,
	}
}

func (m *Monitect) Start(addr string) {
	log.Printf("starting server @ %s", addr)
	m.router.Use(requestLogger)
	m.registerRoutes()
	http.ListenAndServe(addr, m.router)
}

func (m *Monitect) createSensor(w http.ResponseWriter, r *http.Request) {
	var createSensorBody sensor.CreateSensorBody
	if err := json.NewDecoder(r.Body).Decode(&createSensorBody); err != nil {
		common.WriteBody(w, http.StatusBadRequest, &ErrorResponse{Msg: "bad request body", Err: err})
		return
	}
	if sensor, err := m.SensorClient.CreateSensor(&createSensorBody); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "Unable to create sensor", Err: err})
	} else {
		common.WriteBody(w, http.StatusOK, sensor)
	}
}

func (m *Monitect) getSensor(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	if sensor, err := m.SensorClient.GetSensor(sensorId); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "Unable to get sensor", Err: err})
	} else {
		common.WriteBody(w, http.StatusOK, sensor)
	}
}

func (m *Monitect) updateSensor(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	var updateSensorBody sensor.UpdateSensorBody
	if err := json.NewDecoder(r.Body).Decode(&updateSensorBody); err != nil {
		common.WriteBody(w, http.StatusBadRequest, &ErrorResponse{Msg: "bad request body", Err: err})
		return
	}
	if sensor, err := m.SensorClient.UpdateSensor(sensorId, &updateSensorBody); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "Unable to update sensor", Err: err})
	} else {
		common.WriteBody(w, http.StatusOK, sensor)
	}
}

func (m *Monitect) listSensors(w http.ResponseWriter, r *http.Request) {
	if sensors, err := m.SensorClient.ListSensors(); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "Unable to list sensors", Err: err})
	} else {
		common.WriteBody(w, http.StatusOK, &ListSensorsResponse{Sensors: sensors, Count: len(sensors)})
	}
}

func (m *Monitect) deleteSensor(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	if err := m.SensorClient.DeleteSensor(sensorId); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "Unable to delete sensor", Err: err})
	} else {
		common.WriteBody(w, http.StatusOK, common.AnyMap{"msg": "OK"})
	}
}

// publishToStreamFromWebsocket reads the websocket and sends data to the specified stream.Stream
func (m *Monitect) publishToStreamFromWebsocket(ws *websocket.Conn, sensorStream *stream.Manager) {
	defer func() {
		ws.WriteMessage(websocket.CloseMessage, nil)
		ws.Close()
	}()
	for {
		ws.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, d, err := ws.ReadMessage()
		if err != nil {
			return
		}
		sensorStream.Send(d)
	}
}

// publishSensorFeed is used by a sensor to publish new readings
func (m *Monitect) publishSensorFeed(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	sensorStream, found := m.SensorClient.GetSensorStreamManager(sensorId)
	if !found {
		common.WriteBody(w, http.StatusNotFound, &ErrorResponse{Msg: "sensor stream not found " + sensorId})
		return
	}
	// effectively allows cross host
	m.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	if ws, err := m.upgrader.Upgrade(w, r, nil); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "unable to upgrade to websocket", Err: err})
		return
	} else {
		// hand off the processing to another thread so that we can complete this request
		go m.publishToStreamFromWebsocket(ws, sensorStream)
	}
}

// sendReadingsToWebsocket sends messages received on the stream.Stream to the websocket
func (m *Monitect) sendReadingsToWebsocket(ws *websocket.Conn, clientStream *stream.Stream, closeFunc func()) {
	pingTicker := time.NewTicker(2 * time.Second)
	defer func() {
		closeFunc()
		pingTicker.Stop()
		ws.WriteMessage(websocket.CloseMessage, nil)
		ws.Close()
	}()
	for {
		select {
		case data := <-clientStream.C():
			ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("unable to write text message to websocket : %s", err)
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("unable to write ping message to websocket : %s", err)
				return
			}
		}
	}
}

func (m *Monitect) readSensorFeed(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	clientId := r.URL.Query().Get("clientId")
	clientStream, err := m.SensorClient.AddClientStream(sensorId, clientId)
	if err != nil {
		log.Printf("unable to add a client stream: %s", err)
		common.WriteBody(w, http.StatusNotFound, &ErrorResponse{Msg: "sensor not found " + sensorId})
		return
	}
	m.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	if ws, err := m.upgrader.Upgrade(w, r, nil); err != nil {
		log.Printf("unable to upgrade the websocket client: %s", err)
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "unable to upgrade to websocket", Err: err})
		return
	} else {
		// hand off the processing to another thread so that we can complete this request
		log.Printf("starting goroutine for sendReadingsToWebsocket")
		go m.sendReadingsToWebsocket(ws, clientStream, func() { m.SensorClient.RemoveClientStream(sensorId, clientId) })
	}
}

func (m *Monitect) publishSensorReading(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	var sensorReadingBody SensorReadingBody
	if err := json.NewDecoder(r.Body).Decode(&sensorReadingBody); err != nil {
		common.WriteBody(w, http.StatusBadRequest, &ErrorResponse{Msg: "bad request body", Err: err})
		return
	}
	if b, err := json.Marshal(&sensorReadingBody); err != nil {
		common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "unable to json marshal the sensor reading body", Err: err})
		return
	} else {
		if err := m.SensorClient.PublishSensorReading(sensorId, b); err != nil {
			common.WriteBody(w, http.StatusInternalServerError, &ErrorResponse{Msg: "unable to publish sensor reading", Err: err})
			return
		} else {
			common.WriteBody(w, http.StatusOK, common.AnyMap{"msg": "OK"})
			return
		}
	}
}

func main() {
	confPath := os.Args[1]
	log.Printf("using config path %s", confPath)
	// load config
	config := conf.NewConfig(confPath)
	// open the sqlite databa`se, pass it to the models
	db := storage.Connect(config)
	server := NewMonitect(db)
	hostPort := config.Server.Host + ":" + config.Server.Port
	server.Start(hostPort)
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
	var image = document.getElementById("videofeed");
	var updateImage = function(msg) {
		image.src = 'data:image/png;base64, ' + msg.data;
    };

    var ws = new WebSocket("{{.}}");
	ws.onmessage = (event) => {
		updateImage(event);
	};
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Here is an image
<p>
</td><td valign="top" width="50%">
<img id="videofeed">
</img>
</td></tr></table>
</body>
</html>
`))
