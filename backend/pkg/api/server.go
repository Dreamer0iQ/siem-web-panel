package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"siem-project/backend/pkg/storage"
)

type Server struct {
	storage *storage.Storage
	port    int
	server  *http.Server
	users   map[string]string // username -> sha256(password)
}

func NewServer(store *storage.Storage, port int) *Server {
	// Инициализируем пользователей с захешированными паролями
	// В production это должно быть в БД или конфиге
	users := make(map[string]string)

	// Пример: admin:admin123 (в реальной системе пароли должны быть сложнее)
	users["admin"] = hashPassword("admin123")
	users["operator"] = hashPassword("operator123")

	return &Server{
		storage: store,
		port:    port,
		users:   users,
	}
}

// hashPassword создает SHA-256 хэш пароля
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/events", s.corsMiddleware(s.authMiddleware(s.handleEvents)))
	mux.HandleFunc("/api/stats", s.corsMiddleware(s.authMiddleware(s.handleStats)))

	// Health endpoint открыт для мониторинга
	mux.HandleFunc("/api/health", s.corsMiddleware(s.handleHealth))

	mux.HandleFunc("/api/dashboard/agents", s.corsMiddleware(s.authMiddleware(s.handleDashboardAgents)))
	mux.HandleFunc("/api/dashboard/logins", s.corsMiddleware(s.authMiddleware(s.handleDashboardLogins)))
	mux.HandleFunc("/api/dashboard/hosts", s.corsMiddleware(s.authMiddleware(s.handleDashboardHosts)))
	mux.HandleFunc("/api/dashboard/events-by-type", s.corsMiddleware(s.authMiddleware(s.handleDashboardEventsByType)))
	mux.HandleFunc("/api/dashboard/events-by-severity", s.corsMiddleware(s.authMiddleware(s.handleDashboardEventsBySeverity)))
	mux.HandleFunc("/api/dashboard/top-users", s.corsMiddleware(s.authMiddleware(s.handleDashboardTopUsers)))
	mux.HandleFunc("/api/dashboard/top-processes", s.corsMiddleware(s.authMiddleware(s.handleDashboardTopProcesses)))
	mux.HandleFunc("/api/dashboard/timeline", s.corsMiddleware(s.authMiddleware(s.handleDashboardTimeline)))

	// Статика фронтенда
	fs := http.FileServer(http.Dir("./frontend/dist"))
	mux.Handle("/", fs)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="SIEM System"`)
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			http.Error(w, "Unauthorized: invalid Authorization format", http.StatusUnauthorized)
			return
		}

		payload := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			http.Error(w, "Unauthorized: invalid base64 encoding", http.StatusUnauthorized)
			return
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			http.Error(w, "Unauthorized: invalid credentials format", http.StatusUnauthorized)
			return
		}

		username := credentials[0]
		password := credentials[1]

		if !s.validateCredentials(username, password) {
			log.Printf("Failed login attempt for user: %s from %s", username, r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid username or password", http.StatusUnauthorized)
			return
		}

		log.Printf("Successful authentication for user: %s from %s", username, r.RemoteAddr)
		next(w, r)
	}
}

// validateCredentials проверяет username и password
func (s *Server) validateCredentials(username, password string) bool {
	expectedHash, exists := s.users[username]
	if !exists {
		return false
	}

	actualHash := hashPassword(password)

	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) == 1
}

func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// handleEvents обрабатывает запросы к /api/events
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getEvents(w, r)
	case "POST":
		s.addEvents(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getEvents возвращает список событий
func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := storage.EventFilter{
		Source:   query.Get("source"),
		Severity: query.Get("severity"),
		Hostname: query.Get("host"),
		Type:     query.Get("type"),
		User:     query.Get("user"),
		Process:  query.Get("process"),
		From:     query.Get("from"),
		To:       query.Get("to"),
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 20
	}

	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	} else {
		filter.Page = 1
	}

	events, total, err := s.storage.GetEvents(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pages := 0
	if filter.Limit > 0 {
		pages = (total + filter.Limit - 1) / filter.Limit
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  total,
		"page":   filter.Page,
		"limit":  filter.Limit,
		"pages":  pages,
	})
}

func (s *Server) addEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Database   string           `json:"database"`
		Collection string           `json:"collection"`
		Events     []*storage.Event `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Events) == 0 {
		http.Error(w, "No events provided", http.StatusBadRequest)
		return
	}

	if err := s.storage.AddEvents(req.Events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Received %d events from agent", len(req.Events))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Added %d events", len(req.Events)),
		"count":   len(req.Events),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.storage.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	expectedPasswordHash, ok := s.users[req.Username]
	if !ok {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	providedPasswordHash := hashPassword(req.Password)
	if subtle.ConstantTimeCompare([]byte(expectedPasswordHash), []byte(providedPasswordHash)) != 1 {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token := generateToken(req.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
	})
}

func generateToken(username string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, time.Now().Add(24*time.Hour).String())))
}
