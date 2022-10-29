package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr     = flag.String("addr", "localhost:8080", "http service address")
	sensorId = flag.String("sensor", "03017f18-da09-417e-8dc8-5d4109090b11", "the sensor id to read from")
	clientId = flag.String("client", "kenton", "")
)

func readWs(addr string, sensorId string, clientId string) {
	u := "ws://" + addr + "/sensors/" + sensorId + "/feed/read?clientId=" + clientId
	log.Printf("connecting to %s", u)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("%d: received msg: %s", time.Now().UnixMilli(), message)
	}
}

func writeWs(addr string, sensorId string, clientId string) {
	u := "ws://" + addr + "/sensors/" + sensorId + "/feed/publish?clientId=" + clientId
	log.Printf("connecting to %s", u)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		msg := fmt.Sprintf("message:%d", i)
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Printf("write error: %s", err)
		} else {
			log.Printf("%d: wrote msg  : %s", time.Now().UnixMilli(), msg)
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go readWs(*addr, *sensorId, *clientId)
	go writeWs(*addr, *sensorId, *clientId)
	time.Sleep(30 * time.Second)
}
