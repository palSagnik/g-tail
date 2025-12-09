package broker

import (
	"log"
	"sync"
)

// Broker: This is for managing clients, deletion, message update
// Can be understood as a hub type
type Broker struct {
	clients    map[chan string]bool
	NewClients chan chan string
	Defunct    chan chan string
	Message    chan string
	mutex      sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: map[chan string]bool{},
		NewClients: make(chan chan string),
		Defunct: make(chan chan string),
		Message: make(chan string),
	}
}

// This function is for starting a connection
func (b *Broker) Start() {
	for {
		select {
		// 1. New client incoming
		case s := <-b.NewClients:
			b.mutex.Lock()
			b.clients[s] = true
			b.mutex.Unlock()
			log.Println("Broker: New Client Connected")
		
		// 2. Client Disconnection
		case s := <-b.Defunct:
			b.mutex.Lock()
			delete(b.clients, s)
			close(s)
			b.mutex.Unlock()
			log.Println("Broker: Client Disconnected")
		
		// 3. Message update
		case msg := <-b.Message:
			b.mutex.RLock()
			for clientMsgChan := range b.clients {
				select{
				case clientMsgChan <- msg:
				default:
					// Non-blocking send: if client is slow, we skip to prevent blocking
				}
			}
			b.mutex.RUnlock()
		}
	}
}
