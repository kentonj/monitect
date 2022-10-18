package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kentonj/monitect/internal/stream"
)

var upgrader = websocket.Upgrader{}

// publishToStream publishes to the stream,
// if the channel is full, it reads the last message, then republishes
func publishToStream(ch chan []byte, data []byte) {
	select {
	case ch <- data:
	default:
		<-ch
		ch <- data
	}
}

type Server struct {
	websocket.Upgrader
	router        *mux.Router
	streamRPS     time.Duration
	StreamManager *stream.Manager
}

func NewServer(streamManager *stream.Manager) *Server {
	streamRPS := time.Second / 30
	server := Server{
		Upgrader:  websocket.Upgrader{},
		router:    mux.NewRouter(),
		streamRPS: streamRPS, StreamManager: streamManager,
	}
	return &server
}

func (s *Server) registerRoutes() {
	s.router.HandleFunc("/sensors", s.createSensor).Methods(http.MethodPost)
	s.router.HandleFunc("/sensors", s.listSensors).Methods(http.MethodGet)
	s.router.HandleFunc("/sensors/{sensorId}", s.getSensor).Methods(http.MethodGet)
	s.router.HandleFunc("/sensors/{sensorId}/feed", s.readStream).Methods(http.MethodGet)
	s.router.HandleFunc("/sensors/{sensorId}", s.deleteSensor).Methods(http.MethodDelete)

}

func (s *Server) Start(addr string) {
	log.Printf("starting server at %s", addr)
	s.registerRoutes()
	http.ListenAndServe(addr, s.router)
}

func (s *Server) createSensor(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) getSensor(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) listSensors(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) deleteSensor(w http.ResponseWriter, r *http.Request) {
}

// writeJsonResponse writes the statusCode and body. It will panic if it's unable to write
func writeJsonResponse(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if b, err := json.Marshal(body); err != nil {
		panic(err)
	} else {
		if _, err := w.Write(b); err != nil {
			panic(err)
		}
	}
	return
}

type ErrorResponse struct {
	Msg string `json:"msg"`
	Err error  `json:"err"`
}

func NewErrorResponse(msg string, err error) *ErrorResponse {
	return &ErrorResponse{Msg: msg, Err: err}
}

// readStream
// TODO: find the stream for the sensorId specified in the path
func (s *Server) readStream(w http.ResponseWriter, r *http.Request) {
	sensorId := mux.Vars(r)["sensorId"]
	stream, found := s.StreamManager.Stream(sensorId)
	if !found {
		writeJsonResponse(w, http.StatusNotFound,
			NewErrorResponse("did not stream for sensorId "+sensorId, errors.New("sensor stream not found")),
		)
		return
	}
	clientId := r.URL.Query().Get("clientId")
	if err := stream.AddClient(clientId); err != nil {
		writeJsonResponse(w, http.StatusBadRequest,
			NewErrorResponse("unable to add client to stream", err),
		)
		return
	}
	defer stream.RemoveClient(clientId)
	if clientStream, found := stream.Client(clientId); !found {
		writeJsonResponse(w, http.StatusInternalServerError,
			NewErrorResponse("could not locate client stream", errors.New("did not find client stream")),
		)
		return
	} else {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			writeJsonResponse(w, http.StatusInternalServerError,
				NewErrorResponse("unable to upgrade to websocket", errors.New("unable to upgrade to websocket")),
			)
			return
		}
		// read from the stream, but only send data if the ticker has fired (i.e. rate limiting)
		ticker := time.NewTicker(time.Second / 5)
		for msg := range clientStream {
			select {
			case <-ticker.C:
				if err := ws.WriteMessage(1, msg); err != nil {
					return
				}
			default:
			}
		}
	}
}

func main() {
	streamer := stream.NewManager(1000)
	streamer.Add("video")
	stream, _ := streamer.Stream("video")
	go func() {
		log.Print("starting to publish to streamer in the background")
		time.Sleep(5 * time.Second)
		i := 0
		for {
			stream.Publish([]byte(fmt.Sprintf("message: %d", i)))
			time.Sleep(time.Second / 5)
			i += 1
		}
	}()
	server := NewServer(streamer)
	server.Start(":8080")
}
