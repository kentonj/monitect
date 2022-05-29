package common

import (
	"encoding/json"
	"log"
	"net/http"
)

type AnyMap map[string]interface{}

// WriteBody attaches a status code and writes the body as json to the http.ResponseWriter
func WriteBody(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		w.Header().Set("Content-Type", "application/json")
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Panic("unable to marshal the following object as json", body)
		}
		w.Write(jsonBody)
	}
}

func WriteData(w http.ResponseWriter, statusCode int, contentType string, data []byte) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// BindJSON reads the body into the target struct
func BindJSON(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	} else {
		return nil
	}
}
