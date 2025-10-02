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

// StatusUpdate Add this struct near your other data models (User, Message, etc.)
type StatusUpdate struct {
	Type      string `json:"type"` // "status_update"
	UserID    int64  `json:"user_id"`
	NewStatus string `json:"new_status"`
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
//
//	func (h *Hub) Run() {
//		for {
//			select {
//			case client := <-h.Register:
//				h.Clients[client.ID] = client
//				log.Printf("User %d connected, total clients: %d", client.ID, len(h.Clients))
//
//			case client := <-h.Unregister:
//				if _, ok := h.Clients[client.ID]; ok {
//					client.Conn.Close()
//					delete(h.Clients, client.ID)
//					log.Printf("User %d disconnected, total clients: %d", client.ID, len(h.Clients))
//				}
//
//			case message := <-h.Broadcast:
//				// Save message to DB first
//				msgID, err := saveMessage(message)
//				if err != nil {
//					log.Printf("Failed to save message: %v", err)
//					continue
//				}
//				message.ID = msgID
//				loc, _ := time.LoadLocation("Africa/Nairobi")
//				message.CreatedAt = time.Now().In(loc).Format(time.RFC3339)
//				// ISO string
//
//				// Broadcast to recipients
//				for _, uid := range message.RecipientIDs {
//					if c, ok := h.Clients[uid]; ok {
//						if err := c.Conn.WriteJSON(message); err != nil {
//							log.Printf("Error sending message to user %d: %v", uid, err)
//							c.Conn.Close()
//							delete(h.Clients, uid)
//						}
//					}
//				}
//			}
//		}
//	}
//
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
			// Save message to DB and get recipients
			msgID, recipientIDs, err := saveMessage(message) // <-- Capture recipientIDs
			if err != nil {
				log.Printf("Failed to save message: %v", err)
				continue
			}
			message.ID = msgID

			// CRITICAL: Set the recipients before broadcasting
			message.RecipientIDs = recipientIDs

			loc, _ := time.LoadLocation("Africa/Nairobi")
			// Note: The original code re-calculates time here.
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
//

// ==== Save message to DB and fetch recipients ====
func saveMessage(msg Message) (int64, []int64, error) { // <-- Added []int64 return
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
		return 0, nil, err // Return nil for recipients on error
	}
	msgID, _ := res.LastInsertId()

	// --- New/Improved Logic: Fetch all participant IDs ---
	var recipientIDs []int64

	// Fetch all participants for the conversation
	rows, err := db.Query(`
        SELECT user_id, status
        FROM conversation_participants cp
        JOIN users u ON cp.user_id = u.id
        WHERE conversation_id = ?`, msg.ConversationID)
	if err != nil {
		return msgID, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		var status string
		rows.Scan(&uid, &status)

		recipientIDs = append(recipientIDs, uid) // Collect recipient ID

		mStatus := "sent"
		if status == "online" {
			mStatus = "delivered"
		}
		// Note: The original code inserts status for *all* users, including the sender, which is fine.
		db.Exec("INSERT INTO message_status (message_id, user_id, status) VALUES (?, ?, ?)", msgID, uid, mStatus)
	}

	return msgID, recipientIDs, nil // <-- Return the list of recipients
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
	r.HandleFunc("/health", healthHandler).Methods("GET", "OPTIONS")
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/register", registerHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/login", loginHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/logout", logoutHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/users", listUsersHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations", createConversationHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations", listConversationsHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/messages", sendMessageHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/messages", listMessagesHandler).Methods("GET", "OPTIONS")

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

	// In loginHandler, after: u.LastSeen = &now
	// ...
	// update status online + last_seen
	//_, _ = db.Exec("UPDATE users SET status='online', last_seen=NOW() WHERE id=?", u.ID)
	//u.Status = "online"
	//now := time.Now()
	//u.LastSeen = &now

	// --- ADDED: Broadcast Online Status ---
	statusMsg := StatusUpdate{
		Type:      "status_update",
		UserID:    u.ID,
		NewStatus: "online",
	}
	// Broadcast the status change (we'll create a dedicated broadcast channel later)
	// For now, we can write directly to all connected clients in the Hub
	go func() {
		for _, client := range hub.Clients {
			if client.ID != u.ID { // Don't send status update to self (they know they logged in)
				if err := client.Conn.WriteJSON(statusMsg); err != nil {
					log.Printf("Error broadcasting status update to user %d: %v", client.ID, err)
				}
			}
		}
	}()
	// --- END ADDED ---

	// create JWT token
	// ... (rest of function)

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

// Add this StatusUpdate struct somewhere with your other data models (e.g., User, Message)
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int64 `json:"user_id"`
	}

	// Prioritize decoding the user_id from the POST request JSON body
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Fallback check for missing body, not an actual error
			if err.Error() != "EOF" {
				httpError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
		}
	}

	// Safety check: if not found in body, check query param (for testing convenience)
	if req.UserID == 0 {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr != "" {
			req.UserID, _ = strconv.ParseInt(userIDStr, 10, 64)
		}
	}

	if req.UserID == 0 {
		httpError(w, http.StatusBadRequest, "user_id is required in the JSON body or query parameter for logout")
		return
	}

	// 1. Update status to offline and set last_seen in DB
	_, err := db.Exec("UPDATE users SET status='offline', last_seen=NOW() WHERE id=?", req.UserID)
	if err != nil {
		log.Printf("Error during logout status update for user %d: %v", req.UserID, err)
		httpError(w, http.StatusInternalServerError, "DB error during status update")
		return
	}

	// --- START: Broadcast Offline Status to ALL Other Connected Clients ---
	statusMsg := StatusUpdate{
		Type:      "status_update",
		UserID:    req.UserID,
		NewStatus: "offline",
	}

	// Use a goroutine to prevent blocking the HTTP response
	go func() {
		for _, client := range hub.Clients {
			// Only send the status update to other users, not the one logging out
			if client.ID != req.UserID {
				if err := client.Conn.WriteJSON(statusMsg); err != nil {
					log.Printf("Error broadcasting status update to user %d: %v", client.ID, err)
					// If the write fails, assume the connection is dead and clean up
					hub.Unregister <- client
				}
			}
		}
	}()
	// --- END: Broadcast Offline Status ---

	// 2. Forcibly close the WebSocket connection in the Hub
	if client, ok := hub.Clients[req.UserID]; ok {
		// This sends the client to the Unregister channel, which cleans up the Hub map and closes the connection.
		hub.Unregister <- client
	} else {
		log.Printf("User %d was logged out via API but was not found in the Hub clients map.", req.UserID)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully, status set to offline"})
}

// logoutHandler updates user status to 'offline' and invalidates the session (conceptually).
// This expects the user's ID to be passed, typically via a JWT in a real-world app,
// but for simplicity, we'll use a query parameter or body payload.
// logoutHandler updates user status to 'offline' and closes the WebSocket connection.
//func logoutHandler(w http.ResponseWriter, r *http.Request) {
//	var req struct {
//		UserID int64 `json:"user_id"`
//	}
//
//	// Prioritize decoding the user_id from the POST request JSON body
//	if r.Header.Get("Content-Type") == "application/json" {
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			// Fallback check for missing body, not an actual error
//			if err.Error() != "EOF" {
//				httpError(w, http.StatusBadRequest, "invalid JSON body")
//				return
//			}
//		}
//	}
//
//	// Safety check: if not found in body, check query param (for testing convenience)
//	if req.UserID == 0 {
//		userIDStr := r.URL.Query().Get("user_id")
//		if userIDStr != "" {
//			req.UserID, _ = strconv.ParseInt(userIDStr, 10, 64)
//		}
//	}
//
//	if req.UserID == 0 {
//		httpError(w, http.StatusBadRequest, "user_id is required in the JSON body or query parameter for logout")
//		return
//	}
//
//	// 1. Update status to offline and set last_seen in DB
//	_, err := db.Exec("UPDATE users SET status='offline', last_seen=NOW() WHERE id=?", req.UserID)
//	if err != nil {
//		log.Printf("Error during logout status update for user %d: %v", req.UserID, err)
//		httpError(w, http.StatusInternalServerError, "DB error during status update")
//		return
//	}
//
//	// 2. Forcibly close the WebSocket connection in the Hub
//	if client, ok := hub.Clients[req.UserID]; ok {
//		// This sends the client to the Unregister channel, which cleans up the Hub map and closes the connection.
//		hub.Unregister <- client
//		// CRITICAL NOTE: The connection removal is handled in the Hub.Run loop now.
//	} else {
//		log.Printf("User %d was logged out via API but was not found in the Hub clients map.", req.UserID)
//	}
//
//	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully, status set to offline"})
//}

// creating conversations
//func createConversationHandler(w http.ResponseWriter, r *http.Request) {
//	var req createConversationRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		httpError(w, http.StatusBadRequest, "invalid JSON body")
//		return
//	}
//
//	if len(req.ParticipantIDs) < 2 {
//		httpError(w, http.StatusBadRequest, "at least 2 participants required")
//		return
//	}
//
//	// Insert into conversations
//	res, err := db.Exec("INSERT INTO conversations (name, is_group) VALUES (?, ?)",
//		sql.NullString{String: req.Name, Valid: req.IsGroup}, req.IsGroup)
//	if err != nil {
//		httpError(w, http.StatusInternalServerError, "db error: "+err.Error())
//		return
//	}
//
//	convID, _ := res.LastInsertId()
//
//	// Insert participants
//	for _, uid := range req.ParticipantIDs {
//		_, _ = db.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", convID, uid)
//	}
//
//	resp := conversationResponse{
//		ID:             convID,
//		Name:           req.Name,
//		IsGroup:        req.IsGroup,
//		ParticipantIDs: req.ParticipantIDs,
//		CreatedAt:      time.Now(),
//	}
//
//	respondJSON(w, http.StatusCreated, map[string]any{"conversation": resp})
//}

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

	// --- START: Uniqueness Check for 1-on-1 Chats ---
	if !req.IsGroup && len(req.ParticipantIDs) == 2 {
		u1 := req.ParticipantIDs[0]
		u2 := req.ParticipantIDs[1]

		// Ensure u1 is the smaller ID for canonical ordering
		if u1 > u2 {
			u1, u2 = u2, u1
		}

		// Find existing 1-on-1 conversation
		var existingConvID int64
		err := db.QueryRow(`
            SELECT c.id FROM conversations c
            JOIN conversation_participants cp1 ON cp1.conversation_id = c.id AND cp1.user_id = ?
            JOIN conversation_participants cp2 ON cp2.conversation_id = c.id AND cp2.user_id = ?
            WHERE c.is_group = 0 AND 
                  (SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = c.id) = 2
            LIMIT 1
        `, u1, u2).Scan(&existingConvID)

		if err != nil && err != sql.ErrNoRows {
			httpError(w, http.StatusInternalServerError, "db error during uniqueness check: "+err.Error())
			return
		}

		if existingConvID > 0 {
			// Found existing conversation, retrieve and return it
			c := conversationResponse{}
			var name sql.NullString
			err := db.QueryRow("SELECT id, name, is_group, created_at FROM conversations WHERE id = ?", existingConvID).
				Scan(&c.ID, &name, &c.IsGroup, &c.CreatedAt)
			if err == nil {
				if name.Valid {
					c.Name = name.String
				}
				c.ParticipantIDs = req.ParticipantIDs
				respondJSON(w, http.StatusOK, map[string]any{"conversation": c, "message": "Conversation already exists"})
				return
			}
		}
	}
	// --- END: Uniqueness Check ---

	// Insert into conversations (Only runs if no existing 1-on-1 chat was found)
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
