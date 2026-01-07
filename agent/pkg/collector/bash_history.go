// prac3
package collector

import (
	"fmt"
	"strings"

	"siem-project/agent/pkg/types"
)

// парсит ~/.bash_history
type BashHistoryParser struct{}

func NewBashHistoryParser() *BashHistoryParser {
	return &BashHistoryParser{}
}

func (p *BashHistoryParser) GetSourceType() string {
	return "bash_history"
}

func (p *BashHistoryParser) Parse(line string, hostname string) (*types.Event, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	event := types.NewEvent("bash_history", "command_executed", "low", line)
	event.SetHostname(hostname)
	event.Command = line

	// severity по команде
	if isSudoCommand(line) {
		event.Severity = "medium"
		event.EventType = "privileged_command"
	}

	if isDangerousCommand(line) {
		event.Severity = "high"
		event.EventType = "dangerous_command"
	}

	if strings.Contains(line, "sudo su") {
		event.User = "root"
	}

	return event, nil
}

// является ли команда sudo
func isSudoCommand(cmd string) bool {
	return strings.HasPrefix(cmd, "sudo ")
}

// проверяет опасные команды
func isDangerousCommand(cmd string) bool {
	dangerous := []string{
		"rm -rf",
		"dd if=",
		"mkfs",
		"fdisk",
		"passwd root",
		"> /dev/",
		"chmod 777",
	}

	for _, d := range dangerous {
		if strings.Contains(cmd, d) {
			return true
		}
	}
	return false
}
