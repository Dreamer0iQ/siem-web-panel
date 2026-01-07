package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"siem-project/backend/pkg/storage"
)

func (s *Server) handleDashboardAgents(w http.ResponseWriter, r *http.Request) {
	events, _, err := s.storage.GetEvents(storage.EventFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agentsMap := make(map[string]map[string]interface{})
	for _, e := range events {
		if _, exists := agentsMap[e.Host]; !exists {
			agentsMap[e.Host] = map[string]interface{}{
				"hostname":      e.Host,
				"ip_address":    "unknown",
				"last_activity": e.Timestamp,
				"status":        "active",
			}
		} else {
			last := agentsMap[e.Host]["last_activity"].(string)
			if e.Timestamp > last {
				agentsMap[e.Host]["last_activity"] = e.Timestamp
			}
		}
	}

	var agents []map[string]interface{}
	for _, a := range agentsMap {
		agents = append(agents, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// handleDashboardLogins возвращает последние входы
func (s *Server) handleDashboardLogins(w http.ResponseWriter, r *http.Request) {
	events, _, err := s.storage.GetEvents(storage.EventFilter{Limit: 10})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var logins []map[string]interface{}
	for _, e := range events {
		// Очень простой фильтр для демо
		if e.Type == "user_session_start" || e.Type == "login" {
			logins = append(logins, map[string]interface{}{
				"timestamp":  e.Timestamp,
				"user":       e.User,
				"host":       e.Host,
				"success":    true, 
				"ip_address": "unknown",
			})
		}
	}
	if len(logins) == 0 && len(events) > 0 {
		for i, e := range events {
			if i >= 5 { break }
			logins = append(logins, map[string]interface{}{
				"timestamp":  e.Timestamp,
				"user":       e.User,
				"host":       e.Host,
				"success":    true, 
				"ip_address": "unknown",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logins)
}

func (s *Server) handleDashboardHosts(w http.ResponseWriter, r *http.Request) {
	events, _, err := s.storage.GetEvents(storage.EventFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hostsMap := make(map[string]map[string]interface{})
	for _, e := range events {
		if _, exists := hostsMap[e.Host]; !exists {
			hostsMap[e.Host] = map[string]interface{}{
				"hostname":    e.Host,
				"event_count": 0,
				"last_event":  e.Timestamp,
				"ip_address":  "unknown",
			}
		}
		
		h := hostsMap[e.Host]
		h["event_count"] = h["event_count"].(int) + 1
		if e.Timestamp > h["last_event"].(string) {
			h["last_event"] = e.Timestamp
		}
	}

	var hosts []map[string]interface{}
	for _, h := range hostsMap {
		hosts = append(hosts, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hosts)
}

func (s *Server) handleDashboardEventsByType(w http.ResponseWriter, r *http.Request) {
	stats := s.storage.GetStats()
	byType, ok := stats["by_type"].(map[string]int)
	
	var result []map[string]interface{}
	if ok {
		for k, v := range byType {
			result = append(result, map[string]interface{}{
				"type":  k,
				"count": v,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleDashboardEventsBySeverity(w http.ResponseWriter, r *http.Request) {
	stats := s.storage.GetStats()
	bySeverity, ok := stats["by_severity"].(map[string]int)
	
	var result []map[string]interface{}
	if ok {
		for k, v := range bySeverity {
			result = append(result, map[string]interface{}{
				"severity": k,
				"count":    v,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleDashboardTopUsers(w http.ResponseWriter, r *http.Request) {
	events, _, _ := s.storage.GetEvents(storage.EventFilter{})
	userCounts := make(map[string]int)
	
	for _, e := range events {
		if e.User != "" {
			userCounts[e.User]++
		}
	}

	var result []map[string]interface{}
	for u, c := range userCounts {
		result = append(result, map[string]interface{}{
			"username":    u,
			"event_count": c,
		})
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i]["event_count"].(int) > result[j]["event_count"].(int)
	})
	
	if len(result) > 5 {
		result = result[:5]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleDashboardTopProcesses(w http.ResponseWriter, r *http.Request) {
	events, _, _ := s.storage.GetEvents(storage.EventFilter{})
	procCounts := make(map[string]int)
	
	for _, e := range events {
		procName := e.Process
		if procName == "" {
			procName = e.Source
		}
		if procName != "" {
			procCounts[procName]++
		}
	}

	var result []map[string]interface{}
	for p, c := range procCounts {
		result = append(result, map[string]interface{}{
			"process_name": p,
			"event_count":  c,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i]["event_count"].(int) > result[j]["event_count"].(int)
	})

	if len(result) > 5 {
		result = result[:5]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDashboardTimeline возвращает таймлайн событий
func (s *Server) handleDashboardTimeline(w http.ResponseWriter, r *http.Request) {
	events, _, err := s.storage.GetEvents(storage.EventFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timelineMap := make(map[string]int)
	
	lastEventTime := time.Now()
	if len(events) > 0 {
		if t, err := time.Parse(time.RFC3339, events[0].Timestamp); err == nil {
			lastEventTime = t
		}
	}
	
	lastEventTime = lastEventTime.Truncate(time.Hour).Add(time.Hour)
	
	for i := 23; i >= 0; i-- {
		hour := lastEventTime.Add(time.Duration(-i) * time.Hour).Format("15:00")
		timelineMap[hour] = 0
	}

	for _, e := range events {
		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err == nil {
			diff := lastEventTime.Sub(t)
			if diff > 0 && diff <= 24*time.Hour {
				hour := t.Format("15:00")
				if _, ok := timelineMap[hour]; ok {
					timelineMap[hour]++
				}
			}
		}
	}

	var timeline []map[string]interface{}
	for i := 23; i >= 0; i-- {
		hour := lastEventTime.Add(time.Duration(-i) * time.Hour).Format("15:00")
		timeline = append(timeline, map[string]interface{}{
			"hour":  hour,
			"count": timelineMap[hour],
		})
	}
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(timeline)
}
