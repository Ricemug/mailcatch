package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (hub *WebSocketHub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.mutex.Lock()
			hub.clients[client] = true
			hub.mutex.Unlock()
			log.Printf("WebSocket client connected. Total: %d", len(hub.clients))
			
		case client := <-hub.unregister:
			hub.mutex.Lock()
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				client.Close()
			}
			hub.mutex.Unlock()
			log.Printf("WebSocket client disconnected. Total: %d", len(hub.clients))
			
		case message := <-hub.broadcast:
			hub.mutex.RLock()
			for client := range hub.clients {
				select {
				case _ = <-make(chan struct{}):
					// Non-blocking send
				default:
					err := client.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						log.Printf("WebSocket write error: %v", err)
						client.Close()
						delete(hub.clients, client)
					}
				}
			}
			hub.mutex.RUnlock()
		}
	}
}

func (hub *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	hub.register <- conn
	
	// Handle client disconnection
	defer func() {
		hub.unregister <- conn
	}()
	
	// Keep connection alive and handle pings
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		
		if messageType == websocket.CloseMessage {
			break
		}
	}
}

func (hub *WebSocketHub) Broadcast(messageType string, data interface{}) {
	message := WebSocketMessage{
		Type: messageType,
		Data: data,
	}
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		return
	}
	
	select {
	case hub.broadcast <- jsonData:
	default:
		log.Println("WebSocket broadcast channel full, dropping message")
	}
}

func (hub *WebSocketHub) GetClientCount() int {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	return len(hub.clients)
}