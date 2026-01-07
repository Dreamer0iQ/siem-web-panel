// prac3
package collector

import (
	"fmt"
	"regexp"
	"strings"

	"siem-project/agent/pkg/types"
)

// SyslogParser парсит /var/log/syslog и /var/log/auth.log
type SyslogParser struct {
	sourceType string
	lineRegex  *regexp.Regexp
	sudoRegex  *regexp.Regexp
}

func NewSyslogParser(sourceType string) *SyslogParser {
	return &SyslogParser{
		sourceType: sourceType,
		lineRegex: regexp.MustCompile(`^(\S+)\s+(\S+)\s+([^:\[\s]+)(?:\[(\d+)\])?\s*:\s*(.+)$`),
		// Для sudo логов: parallels : TTY=pts/0 ; PWD=/home/parallels ; USER=root ; COMMAND=/usr/bin/tail
		sudoRegex: regexp.MustCompile(`(\w+)\s*:.*USER=(\w+)\s*;\s*COMMAND=(.+)$`),
	}
}

func (p *SyslogParser) GetSourceType() string {
	return p.sourceType
}

func (p *SyslogParser) Parse(line string, hostname string) (*types.Event, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	matches := p.lineRegex.FindStringSubmatch(line)
	if len(matches) < 6 {
		return nil, fmt.Errorf("failed to parse syslog line")
	}

	process := matches[3]
	pid := matches[4]
	message := matches[5]

	event := types.NewEvent(p.sourceType, "system_event", "low", line)
	event.SetHostname(hostname)
	event.Process = process
	if pid != "" {
		event.Process = fmt.Sprintf("%s[%s]", process, pid)
	}

	p.classifyEvent(event, process, message)

	if process == "sudo" && strings.Contains(message, "COMMAND=") {
		p.parseSudoLog(event, message)
	}

	return event, nil
}

func (p *SyslogParser) classifyEvent(event *types.Event, process, message string) {
	// Auth события
	if strings.Contains(message, "session opened") {
		event.EventType = "session_opened"
		event.Severity = "medium"
	} else if strings.Contains(message, "session closed") {
		event.EventType = "session_closed"
		event.Severity = "low"
	} else if strings.Contains(message, "authentication failure") {
		event.EventType = "auth_failure"
		event.Severity = "high"
	} else if strings.Contains(message, "Accepted password") || strings.Contains(message, "Accepted publickey") {
		event.EventType = "user_login"
		event.Severity = "medium"
	} else if strings.Contains(message, "Failed password") {
		event.EventType = "login_failed"
		event.Severity = "high"
	}

	// Sudo команды
	if process == "sudo" {
		event.EventType = "sudo_command"
		event.Severity = "medium"

		if containsDangerousCommand(message) {
			event.Severity = "high"
			event.EventType = "dangerous_sudo_command"
		}
	}

	// Системные события
	if process == "systemd" {
		event.EventType = "systemd_event"
		event.Severity = "low"
	}

	// SSH события
	if process == "sshd" {
		event.EventType = "ssh_event"
		event.Severity = "medium"
	}
}

func (p *SyslogParser) parseSudoLog(event *types.Event, message string) {
	matches := p.sudoRegex.FindStringSubmatch(message)
	if len(matches) >= 4 {
		event.User = matches[2]    // USER=root
		event.Command = matches[3] // COMMAND=...
	}
}

func containsDangerousCommand(message string) bool {
	dangerous := []string{
		"rm -rf",
		"dd if=",
		"mkfs",
		"fdisk",
		"passwd",
		"userdel",
		"shutdown",
		"reboot",
		"halt",
	}

	for _, d := range dangerous {
		if strings.Contains(message, d) {
			return true
		}
	}
	return false
}
