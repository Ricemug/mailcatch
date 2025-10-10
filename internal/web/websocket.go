package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

	// Configure connection
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	// Channel to signal reader goroutine exit
	done := make(chan struct{})

	// Start reader goroutine
	go func() {
		defer close(done)
		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				break
			}

			if messageType == websocket.CloseMessage {
				break
			}
		}
	}()

	// Send pings
	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
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