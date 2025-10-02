package main

import (
	"database/sql"
	"encoding/json"
	_ "fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

// === Configuration (change via environment) ===
// Default DSN for XAMPP on Windows (root without password):
//
//	root:@tcp(127.0.0.1:3306)/chat_app?parseTime=true
//
// If root has a password or you use another user, set CHAT_DSN env var.
var DSN = getEnv("CHAT_DSN", "root:@tcp(127.0.0.1:3306)/chat_app?parseTime=true&loc=UTC")
var ADDR = getEnv("CHAT_ADDR", ":8080")
var jwtKey = []byte("my_secret_key")

var db *sql.DB
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for testing
	},
}

// User model (matches chat_app.users exactly)
type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Status    string     `json:"status"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// request payload for register
type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type createConversationRequest struct {
	ParticipantIDs []int64 `json:"participant_ids"`
	Name           string  `json:"name"`
	IsGroup        bool    `json:"is_group"`
}

type conversationResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	IsGroup        bool      `json:"is_group"`
	ParticipantIDs []int64   `json:"participant_ids"`
	CreatedAt      time.Time `json:"created_at"`
}

type sendMessageRequest struct {
	ConversationID int64  `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	Content        string `json:"content"`
	MessageType    string `json:"message_type"` // text, image, video, file
}

type messageResponse struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	CreatedAt      time.Time `json:"created_at"`
}

type Client struct {
	ID   int64
	Conn *websocket.Conn
}

type Hub struct {
	Clients    map[int64]*Client // userID -> client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

type Message struct {
	ID             int64   `json:"id"`
	ConversationID int64   `json:"conversation_id"`
	SenderID       int64   `json:"sender_id"`
	Content        string  `json:"content"`
	MessageType    string  `json:"message_type"`
	RecipientIDs   []int64 `json:"recipient_ids"`
	CreatedAt      string  `json:"created_at"`
}

var hub = Hub{
	Clients:    make(map[int64]*Client),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Broadcast:  make(chan Message),
}

// ==== Hub run loop ====
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ID] = client
			log.Printf("User %d connected, total clients: %d", client.ID, len(h.Clients))

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ID]; ok {
				client.Conn.Close()
				delete(h.Clients, client.ID)
				log.Printf("User %d disconnected, total clients: %d", client.ID, len(h.Clients))
			}

		case message := <-h.Broadcast:
			// Save message to DB first
			msgID, err := saveMessage(message)
			if err != nil {
				log.Printf("Failed to save message: %v", err)
				continue
			}
			message.ID = msgID
			loc, _ := time.LoadLocation("Africa/Nairobi")
			message.CreatedAt = time.Now().In(loc).Format(time.RFC3339)
			// ISO string

			// Broadcast to recipients
			for _, uid := range message.RecipientIDs {
				if c, ok := h.Clients[uid]; ok {
					if err := c.Conn.WriteJSON(message); err != nil {
						log.Printf("Error sending message to user %d: %v", uid, err)
						c.Conn.Close()
						delete(h.Clients, uid)
					}
				}
			}
		}
	}
}

// ==== Save message to DB ====
func saveMessage(msg Message) (int64, error) {
	// Convert CreatedAt string to time.Time
	createdAt, err := time.Parse(time.RFC3339, msg.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}

	res, err := db.Exec(
		"INSERT INTO messages (conversation_id, sender_id, content, message_type, created_at) VALUES (?, ?, ?, ?, ?)",
		msg.ConversationID, msg.SenderID, msg.Content, msg.MessageType, createdAt,
	)
	if err != nil {
		return 0, err
	}
	msgID, _ := res.LastInsertId()

	// Save message status for participants
	rows, err := db.Query(`
        SELECT user_id, status
        FROM conversation_participants cp
        JOIN users u ON cp.user_id = u.id
        WHERE conversation_id = ?`, msg.ConversationID)
	if err != nil {
		return msgID, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		var status string
		rows.Scan(&uid, &status)
		mStatus := "sent"
		if status == "online" {
			mStatus = "delivered"
		}
		db.Exec("INSERT INTO message_status (message_id, user_id, status) VALUES (?, ?, ?)", msgID, uid, mStatus)
	}

	return msgID, nil
}

// ==== WebSocket handler ====
func wsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{ID: userID, Conn: conn}
	hub.Register <- client
	defer func() { hub.Unregister <- client }()

	// Send initial message
	conn.WriteJSON(map[string]string{"message": "connected to chat server"})

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket closed for user %d: %v", userID, err)
			}
			break
		}

		// Set timestamp in ISO string for DB
		loc, _ := time.LoadLocation("Africa/Nairobi")
		msg.CreatedAt = time.Now().In(loc).Format(time.RFC3339)
		hub.Broadcast <- msg

	}
}

// CORS Middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// open DB
	var err error
	db, err = sql.Open("mysql", DSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	// small pool tuning
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	// ping DB
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("connected to DB (chat_app)")

	go hub.Run()
	// router
	r := mux.NewRouter()
	r.Use(corsMiddleware)
	r.HandleFunc("/health", healthHandler).Methods("GET")
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/register", registerHandler).Methods("POST")
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/users", listUsersHandler).Methods("GET")
	api.HandleFunc("/conversations", createConversationHandler).Methods("POST")
	api.HandleFunc("/conversations", listConversationsHandler).Methods("GET")
	api.HandleFunc("/messages", sendMessageHandler).Methods("POST")
	api.HandleFunc("/messages", listMessagesHandler).Methods("GET")

	r.HandleFunc("/ws", wsHandler)

	// start server
	log.Printf("server listening on %s\n", ADDR)
	if err := http.ListenAndServe(ADDR, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 {
		httpError(w, http.StatusBadRequest, "username must be at least 3 characters")
		return
	}
	if len(req.Password) < 6 {
		httpError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// hash password
	pwHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// insert into users (columns: username, password_hash)
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", req.Username, string(pwHash))
	if err != nil {
		// detect duplicate username (MySQL error 1062)
		if me, ok := err.(*mysqlDriver.MySQLError); ok && me.Number == 1062 {
			httpError(w, http.StatusConflict, "username already exists")
			return
		}
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	id, _ := res.LastInsertId()

	// fetch inserted user to return created_at and default fields
	u := User{}
	err = db.QueryRow("SELECT id, username, status, last_seen, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Username, &u.Status, &u.LastSeen, &u.CreatedAt)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to read created user: "+err.Error())
		return
	}

	uResp := map[string]any{"user": u}
	respondJSON(w, http.StatusCreated, uResp)
}

// login Logic
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		httpError(w, http.StatusBadRequest, "username and password required")
		return
	}

	// fetch user
	u := User{}
	var passwordHash string
	err := db.QueryRow("SELECT id, username, password_hash, status, last_seen, created_at FROM users WHERE username = ?", req.Username).
		Scan(&u.ID, &u.Username, &passwordHash, &u.Status, &u.LastSeen, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			httpError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	// verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		httpError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// update status online + last_seen
	_, _ = db.Exec("UPDATE users SET status='online', last_seen=NOW() WHERE id=?", u.ID)
	u.Status = "online"
	now := time.Now()
	u.LastSeen = &now

	// create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString(jwtKey)

	respondJSON(w, http.StatusOK, loginResponse{User: u, Token: tokenStr})
}

// listing users
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, username, status, last_seen, created_at FROM users ORDER BY username ASC")
	if err != nil {
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u := User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Status, &u.LastSeen, &u.CreatedAt); err != nil {
			httpError(w, http.StatusInternalServerError, "scan error: "+err.Error())
			return
		}
		users = append(users, u)
	}

	respondJSON(w, http.StatusOK, map[string]any{"users": users})
}

// creating conversations
func createConversationHandler(w http.ResponseWriter, r *http.Request) {
	var req createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if len(req.ParticipantIDs) < 2 {
		httpError(w, http.StatusBadRequest, "at least 2 participants required")
		return
	}

	// Insert into conversations
	res, err := db.Exec("INSERT INTO conversations (name, is_group) VALUES (?, ?)",
		sql.NullString{String: req.Name, Valid: req.IsGroup}, req.IsGroup)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	convID, _ := res.LastInsertId()

	// Insert participants
	for _, uid := range req.ParticipantIDs {
		_, _ = db.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", convID, uid)
	}

	resp := conversationResponse{
		ID:             convID,
		Name:           req.Name,
		IsGroup:        req.IsGroup,
		ParticipantIDs: req.ParticipantIDs,
		CreatedAt:      time.Now(),
	}

	respondJSON(w, http.StatusCreated, map[string]any{"conversation": resp})
}

// listing conversation
func listConversationsHandler(w http.ResponseWriter, r *http.Request) {
	// For simplicity, take user_id as query param for now
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		httpError(w, http.StatusBadRequest, "user_id required")
		return
	}

	rows, err := db.Query(`
		SELECT c.id, c.name, c.is_group
		FROM conversations c
		JOIN conversation_participants cp ON cp.conversation_id = c.id
		WHERE cp.user_id = ?`, userID)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	defer rows.Close()

	convs := []conversationResponse{}
	for rows.Next() {
		var c conversationResponse
		var name sql.NullString
		if err := rows.Scan(&c.ID, &name, &c.IsGroup); err != nil {
			httpError(w, http.StatusInternalServerError, "scan error: "+err.Error())
			return
		}
		if name.Valid {
			c.Name = name.String
		} else {
			c.Name = ""
		}

		// fetch participant ids
		pRows, _ := db.Query("SELECT user_id FROM conversation_participants WHERE conversation_id = ?", c.ID)
		var pids []int64
		for pRows.Next() {
			var pid int64
			pRows.Scan(&pid)
			pids = append(pids, pid)
		}
		pRows.Close()
		c.ParticipantIDs = pids

		convs = append(convs, c)
	}

	respondJSON(w, http.StatusOK, map[string]any{"conversations": convs})
}

// sending messages
func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// insert message
	res, err := db.Exec("INSERT INTO messages (conversation_id, sender_id, content, message_type) VALUES (?, ?, ?, ?)",
		req.ConversationID, req.SenderID, req.Content, req.MessageType)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	msgID, _ := res.LastInsertId()

	// get participants
	rows, _ := db.Query("SELECT user_id, status FROM conversation_participants cp JOIN users u ON cp.user_id = u.id WHERE conversation_id = ?", req.ConversationID)
	for rows.Next() {
		var uid int64
		var status string
		rows.Scan(&uid, &status)

		msgStatus := "sent"
		if status == "online" {
			msgStatus = "delivered"
		}

		db.Exec("INSERT INTO message_status (message_id, user_id, status) VALUES (?, ?, ?)", msgID, uid, msgStatus)
	}
	rows.Close()

	resp := messageResponse{
		ID:             msgID,
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		CreatedAt:      time.Now(),
	}

	respondJSON(w, http.StatusCreated, map[string]any{"message": resp})
}

// fetching messages
func listMessagesHandler(w http.ResponseWriter, r *http.Request) {
	convID := r.URL.Query().Get("conversation_id")
	if convID == "" {
		httpError(w, http.StatusBadRequest, "conversation_id required")
		return
	}

	rows, err := db.Query("SELECT id, conversation_id, sender_id, content, message_type, created_at FROM messages WHERE conversation_id=? ORDER BY created_at ASC", convID)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	defer rows.Close()

	msgs := []messageResponse{}
	for rows.Next() {
		var m messageResponse
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.MessageType, &m.CreatedAt); err != nil {
			httpError(w, http.StatusInternalServerError, err.Error())
			return
		}
		msgs = append(msgs, m)
	}

	respondJSON(w, http.StatusOK, map[string]any{"messages": msgs})
}

// ==== helpers ====
func httpError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
