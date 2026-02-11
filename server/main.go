package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var jwtSecret = []byte("super_secret_key_change_me")

var db *gorm.DB

// Thread-safe client registry
type Client struct {
	Conn     *websocket.Conn
	Username string
	UserID   uint
}

var (
	clients   = make(map[*websocket.Conn]*Client)
	clientsMu sync.RWMutex
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Host == "localhost:8080"
	},
}

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
}

type Message struct {
	ID        uint `gorm:"primaryKey"`
	SenderID  uint
	Content   string
	Timestamp time.Time
}

// ================= DATABASE =================

func initDB() {
	dsn := "host=localhost user=postgres password=postgres dbname=chatapp port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	db.AutoMigrate(&User{}, &Message{})
}

// ================= JWT =================

func generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", err
	}

	return username, nil
}

// ================= HANDLERS =================

func authHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// Create user if not exists
	var user User
	db.FirstOrCreate(&user, User{Username: username})

	token, err := generateJWT(username)
	if err != nil {
		http.Error(w, "Token generation failed", 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	var list []string
	for _, c := range clients {
		list = append(list, c.Username)
	}

	json.NewEncoder(w).Encode(list)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// First message must contain token
	var authMsg struct {
		Token string `json:"token"`
	}

	err = conn.ReadJSON(&authMsg)
	if err != nil {
		conn.Close()
		return
	}

	username, err := validateJWT(authMsg.Token)
	if err != nil {
		conn.Close()
		return
	}

	// Fetch user
	var user User
	db.First(&user, "username = ?", username)

	client := &Client{
		Conn:     conn,
		Username: username,
		UserID:   user.ID,
	}

	clientsMu.Lock()
	clients[conn] = client
	clientsMu.Unlock()

	broadcastSystem(username + " joined the chat")

	for {
		var msg struct {
			Content string `json:"content"`
		}

		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Save to DB
		db.Create(&Message{
			SenderID:  client.UserID,
			Content:   msg.Content,
			Timestamp: time.Now(),
		})

		broadcastMessage(client.Username, msg.Content)
	}

	// Remove on disconnect
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()

	broadcastSystem(username + " left the chat")
	conn.Close()
}

// ================= BROADCAST =================

func broadcastMessage(sender, content string) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	message := map[string]string{
		"type":    "message",
		"sender":  sender,
		"content": content,
	}

	for _, c := range clients {
		c.Conn.WriteJSON(message)
	}
}

func broadcastSystem(content string) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	message := map[string]string{
		"type":    "system",
		"content": content,
	}

	for _, c := range clients {
		c.Conn.WriteJSON(message)
	}
}
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5500")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ================= MAIN =================

func main() {
	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", authHandler)
	mux.HandleFunc("/ws", wsHandler)
	mux.HandleFunc("/users", usersHandler)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
