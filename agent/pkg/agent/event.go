package agent

import (
	"time"
)

type Event struct {
	Timestamp string `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Source    string `json:"source"`     // auditd, syslog, auth, bash_history
	EventType string `json:"event_type"` // user_login, command, file_access, etc
	Severity  string `json:"severity"`   // low, medium, high, critical
	User      string `json:"user,omitempty"`
	Process   string `json:"process,omitempty"`
	Command   string `json:"command,omitempty"`
	RawLog    string `json:"raw_log"`
}

func NewEvent(source, eventType, severity, rawLog string) *Event {
	return &Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Source:    source,
		EventType: eventType,
		Severity:  severity,
		RawLog:    rawLog,
	}
}

func (e *Event) SetHostname(hostname string) {
	e.Hostname = hostname
}

type Message struct {
	AgentID   string   `json:"agent_id"`
	Timestamp string   `json:"timestamp"`
	Events    []*Event `json:"events"`
}

func NewMessage(agentID string, events []*Event) *Message {
	return &Message{
		AgentID:   agentID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Events:    events,
	}
}
