package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"ghostwire/backend/internal/ws"
	"github.com/google/uuid"
)

type Server struct {
	db        *sql.DB
	hub       *ws.Hub
	staticDir string
}

func New(db *sql.DB, hub *ws.Hub, staticDir string) *Server {
	return &Server{db: db, hub: hub, staticDir: staticDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", s.register)
	mux.HandleFunc("/api/user/", s.user)
	mux.HandleFunc("/api/invite/create", s.createInvite)
	mux.HandleFunc("/api/invite/accept", s.acceptInvite)
	mux.HandleFunc("/ws", s.hub.ServeWS)
	mux.Handle("/", spaFileServer(s.staticDir))
	return metadataStripper(mux)
}

func metadataStripper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Forwarded")
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-Real-IP")
		r.Header.Del("Referer")
		r.Header.Del("User-Agent")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID          string `json:"id"`
		PublicKey   string `json:"public_key"`
		IsEphemeral bool   `json:"is_ephemeral"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" || req.PublicKey == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	_, err := s.db.Exec(`INSERT OR REPLACE INTO users(id, public_key, is_ephemeral, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, req.ID, req.PublicKey, boolInt(req.IsEphemeral))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"id": req.ID})
}

func (s *Server) user(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	var pub string
	var eph int
	if err := s.db.QueryRow(`SELECT public_key, is_ephemeral FROM users WHERE id = ?`, id).Scan(&pub, &eph); err != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, map[string]any{"id": id, "public_key": pub, "is_ephemeral": eph == 1})
}

func (s *Server) createInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		CreatorID string `json:"creator_id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.CreatorID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	token := uuid.NewString()
	expires := time.Now().UTC().Add(24 * time.Hour)
	_, err := s.db.Exec(`INSERT INTO invites(token, creator_id, expires_at) VALUES (?, ?, ?)`, token, req.CreatorID, expires.Format(time.RFC3339))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"token": token, "url": "/invite/" + token})
}

func (s *Server) acceptInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Token == "" || req.UserID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var creator string
	if err := s.db.QueryRow(`SELECT creator_id FROM invites WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`, req.Token).Scan(&creator); err != nil {
		http.Error(w, "invite not found", http.StatusNotFound)
		return
	}
	a, b := ordered(creator, req.UserID)
	_, err := s.db.Exec(`INSERT OR REPLACE INTO contacts(user_a, user_b, status) VALUES (?, ?, 'accepted')`, a, b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "accepted", "user_a": a, "user_b": b})
}

func spaFileServer(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fs.ServeHTTP(w, r) })
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func ordered(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}
