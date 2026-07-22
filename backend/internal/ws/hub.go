package ws

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Event struct {
	Type     string `json:"type"`
	ID       string `json:"id,omitempty"`
	RoomID   string `json:"room_id,omitempty"`
	SenderID string `json:"sender_id,omitempty"`
	Payload  string `json:"payload,omitempty"`
}

type Hub struct {
	db         *sql.DB
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan Event
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{db: db, clients: map[*Client]bool{}, register: make(chan *Client), unregister: make(chan *Client), broadcast: make(chan Event, 256)}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if h.clients[c] {
				delete(h.clients, c)
				close(c.send)
			}
		case e := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- e:
				default:
					close(c.send)
					delete(h.clients, c)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool { return true }}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade:", err)
		return
	}
	client := &Client{hub: h, conn: conn, send: make(chan Event, 64)}
	h.register <- client
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() { c.hub.unregister <- c; c.conn.Close() }()
	c.conn.SetReadLimit(64 * 1024)
	for {
		var e Event
		if err := c.conn.ReadJSON(&e); err != nil {
			return
		}
		switch e.Type {
		case "message":
			e.ID = uuid.NewString()
			if e.RoomID == "" {
				e.RoomID = "global"
			}
			expires := time.Now().UTC().Add(24 * time.Hour)
			_, err := c.hub.db.Exec(`INSERT INTO messages(id, room_id, sender_id, payload, expires_at) VALUES (?, ?, ?, ?, ?)`, e.ID, e.RoomID, e.SenderID, e.Payload, expires.Format(time.RFC3339))
			if err == nil {
				c.hub.broadcast <- e
			}
		case "read_ack":
			_, _ = c.hub.db.Exec(`UPDATE messages SET read_at = CURRENT_TIMESTAMP WHERE id = ? AND read_at IS NULL`, e.ID)
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for e := range c.send {
		b, _ := json.Marshal(e)
		if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
			return
		}
	}
}
