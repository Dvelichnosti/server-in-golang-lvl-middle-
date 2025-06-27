package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/http-swagger"
	_ "goncord/docs"
)

// @title Goncord API
// @version 1.0
// @description Simple real-time sync server
// @host localhost:8080
// @BasePath /

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var jwtKey = []byte("secret")
var redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

var (
	connectionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "goncord_connections", Help: "Connections per room"},
		[]string{"room"},
	)
	messagesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "goncord_messages_total", Help: "Total messages per room"},
		[]string{"room"},
	)
)

func init() {
	prometheus.MustRegister(connectionsGauge, messagesCounter)
}

type Room struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	name       string
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	authUser string
	room     *Room
}

type Message struct {
	UserID  string `json:"user_id"`
	Room    string `json:"room"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}

var rooms = make(map[string]*Room)
var roomsMutex sync.RWMutex

func getOrCreateRoom(name string) *Room {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()
	if room, ok := rooms[name]; ok {
		return room
	}
	room := &Room{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		name:       name,
	}
	rooms[name] = room
	go room.run()
	return room
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.mutex.Lock()
			r.clients[client] = true
			r.mutex.Unlock()
			connectionsGauge.WithLabelValues(r.name).Inc()
		case client := <-r.unregister:
			r.mutex.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
				connectionsGauge.WithLabelValues(r.name).Dec()
			}
			r.mutex.Unlock()
		case message := <-r.broadcast:
			r.saveToRedis(message)
			messagesCounter.WithLabelValues(r.name).Inc()
			r.mutex.RLock()
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
			r.mutex.RUnlock()
		}
	}
}

func (r *Room) saveToRedis(message []byte) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s:messages", r.name)
	redisClient.RPush(ctx, key, message)
	redisClient.LTrim(ctx, key, -100, -1)
}

func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			continue
		}
		m.UserID = c.authUser
		m.Room = c.room.name
		m.Time = time.Now().Unix()
		final, _ := json.Marshal(m)
		c.room.broadcast <- final
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func validateJWT(t string) (string, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	return claims["user_id"].(string), nil
}

func serveWs(w http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")
	roomName := req.URL.Query().Get("room")
	if roomName == "" {
		roomName = "default"
	}
	userID, err := validateJWT(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	room := getOrCreateRoom(roomName)
	client := &Client{conn: conn, send: make(chan []byte, 256), authUser: userID, room: room}
	room.register <- client
	go client.writePump()
	client.readPump()
}

// @Summary Get chat history
// @Produce json
// @Param room query string false "Room name"
// @Success 200 {array} string
// @Router /history [get]
func historyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	key := fmt.Sprintf("room:%s:messages", room)
	messages, err := redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func main() {
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/history", historyHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/swagger/", httpSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Goncord running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
