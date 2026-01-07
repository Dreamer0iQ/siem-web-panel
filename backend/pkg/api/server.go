package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"siem-project/backend/pkg/storage"
)

type Server struct {
	storage *storage.Storage
	port    int
	server  *http.Server
}

func NewServer(store *storage.Storage, port int) *Server {
	return &Server{
		storage: store,
		port:    port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API эндпоинты
	mux.HandleFunc("/api/events", s.corsMiddleware(s.handleEvents))
	mux.HandleFunc("/api/stats", s.corsMiddleware(s.handleStats))
	mux.HandleFunc("/api/health", s.corsMiddleware(s.handleHealth))

	// Dashboard эндпоинты
	mux.HandleFunc("/api/dashboard/agents", s.corsMiddleware(s.handleDashboardAgents))
	mux.HandleFunc("/api/dashboard/logins", s.corsMiddleware(s.handleDashboardLogins))
	mux.HandleFunc("/api/dashboard/hosts", s.corsMiddleware(s.handleDashboardHosts))
	mux.HandleFunc("/api/dashboard/events-by-type", s.corsMiddleware(s.handleDashboardEventsByType))
	mux.HandleFunc("/api/dashboard/events-by-severity", s.corsMiddleware(s.handleDashboardEventsBySeverity))
	mux.HandleFunc("/api/dashboard/top-users", s.corsMiddleware(s.handleDashboardTopUsers))
	mux.HandleFunc("/api/dashboard/top-processes", s.corsMiddleware(s.handleDashboardTopProcesses))
	mux.HandleFunc("/api/dashboard/timeline", s.corsMiddleware(s.handleDashboardTimeline))

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
