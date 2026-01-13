package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"siem-project/backend/pkg/storage"

	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	storage   *storage.Storage
	port      int
	server    *http.Server
	users     map[string]string // username -> sha256(password)
	jwtSecret []byte            // JWT signing key
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewServer(store *storage.Storage, port int) *Server {
	users := make(map[string]string)

	users["admin"] = hashPassword("admin123")
	users["operator"] = hashPassword("operator123")

	jwtSecret := []byte("your-super-secret-jwt-key-change-in-production")

	return &Server{
		storage:   store,
		port:      port,
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Публичные эндпоинты
	mux.HandleFunc("/api/login", s.corsMiddleware(s.handleLogin))
	mux.HandleFunc("/api/health", s.corsMiddleware(s.handleHealth))

	// Защищенные эндпоинты (требуют JWT)
	mux.HandleFunc("/api/events", s.corsMiddleware(s.authMiddleware(s.handleEvents)))
	mux.HandleFunc("/api/stats", s.corsMiddleware(s.authMiddleware(s.handleStats)))

	mux.HandleFunc("/api/dashboard/agents", s.corsMiddleware(s.authMiddleware(s.handleDashboardAgents)))
	mux.HandleFunc("/api/dashboard/logins", s.corsMiddleware(s.authMiddleware(s.handleDashboardLogins)))
	mux.HandleFunc("/api/dashboard/hosts", s.corsMiddleware(s.authMiddleware(s.handleDashboardHosts)))
	mux.HandleFunc("/api/dashboard/events-by-type", s.corsMiddleware(s.authMiddleware(s.handleDashboardEventsByType)))
	mux.HandleFunc("/api/dashboard/events-by-severity", s.corsMiddleware(s.authMiddleware(s.handleDashboardEventsBySeverity)))
	mux.HandleFunc("/api/dashboard/top-users", s.corsMiddleware(s.authMiddleware(s.handleDashboardTopUsers)))
	mux.HandleFunc("/api/dashboard/top-processes", s.corsMiddleware(s.authMiddleware(s.handleDashboardTopProcesses)))
	mux.HandleFunc("/api/dashboard/timeline", s.corsMiddleware(s.authMiddleware(s.handleDashboardTimeline)))

	// Эндпоинт для агента (без JWT аутентификации)
	mux.HandleFunc("/query", s.corsMiddleware(s.handleAgentIngest))

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
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized: invalid Authorization format (expected Bearer token)", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.jwtSecret, nil
		})

		if err != nil {
			log.Printf("JWT parsing error: %v", err)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized: token is not valid", http.StatusUnauthorized)
			return
		}

		// Проверяем, что пользователь существует
		if _, exists := s.users[claims.Username]; !exists {
			log.Printf("Token valid but user not found: %s", claims.Username)
			http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
			return
		}

		log.Printf("Successful JWT authentication for user: %s from %s", claims.Username, r.RemoteAddr)

		// Добавляем username в контекст (можно использовать в handlers)
		// r = r.WithContext(context.WithValue(r.Context(), "username", claims.Username))

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

// handleAgentIngest обрабатывает входящие события от агента (без JWT)
func (s *Server) handleAgentIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Database   string           `json:"database"`
		Collection string           `json:"collection"`
		Events     []*storage.Event `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Agent ingest error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Events) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "No events to add",
			"count":   0,
		})
		return
	}

	if err := s.storage.AddEvents(req.Events); err != nil {
		log.Printf("Agent ingest storage error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Agent ingested %d events from %s/%s", len(req.Events), req.Database, req.Collection)

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

	// Валидация credentials
	if !s.validateCredentials(req.Username, req.Password) {
		log.Printf("Failed login attempt for user: %s from %s", req.Username, r.RemoteAddr)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := s.generateJWT(req.Username)
	if err != nil {
		log.Printf("Error generating JWT for user %s: %v", req.Username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Successful login for user: %s from %s", req.Username, r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
		"user":    req.Username,
	})
}

// generateJWT создает JWT токен для пользователя
func (s *Server) generateJWT(username string) (string, error) {
	// Токен действителен 24 часа
	expirationTime := time.Now().Add(24 * time.Hour)

	// Создаем claims
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "siem-backend",
		},
	}

	// Создаем токен с алгоритмом HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
