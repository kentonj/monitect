package stream

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestManager_SendWithMultipleClients(t *testing.T) {
	sm := NewManager(5)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			msg := fmt.Sprintf("message: %d", i)
			sm.Send([]byte(msg))
			time.Sleep(1 * time.Millisecond)
		}
	}()
	// add 10 clients that are reading from the client streams
	for i := 0; i < 10; i++ {
		clientId := fmt.Sprintf("client:%d", i)
		sm.AddClient(clientId)
		wg.Add(1)
		// start a listener to the internal channel
		go func(clientId string) {
			defer wg.Done()
			clientStream, _ := sm.ClientStream(clientId)
			for msg := range clientStream.C() {
				log.Printf("client %s got message: %s", clientId, msg)
			}
			log.Printf("the channel is now closed for client %s!", clientId)
		}(clientId)
		// remove the client after a random amount of time
		go func(clientId string) {
			sleepDur := time.Duration(rand.Int63n(1000)) * time.Millisecond
			time.Sleep(sleepDur)
			sm.RemoveClient(clientId)
		}(clientId)
	}
	wg.Wait()
}
