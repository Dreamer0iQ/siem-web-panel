package collector

import (
	"regexp"
	"strings"

	"siem-project/agent/pkg/types"
)

// /var/log/audit/audit.log
type AuditdParser struct {
	typeRegex  *regexp.Regexp
	fieldRegex *regexp.Regexp
}

func NewAuditdParser() *AuditdParser {
	return &AuditdParser{
		typeRegex:  regexp.MustCompile(`type=(\w+)`),
		fieldRegex: regexp.MustCompile(`(\w+)=([^\s]+)`),
	}
}

func (p *AuditdParser) GetSourceType() string {
	return "auditd"
}

func (p *AuditdParser) Parse(line string, hostname string) (*types.Event, error) {
	// type=SYSCALL msg=audit(1234567890.123:456): arch=c000003e syscall=59 success=yes
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	auditType := p.extractType(line)
	eventType := p.mapAuditType(auditType)
	severity := p.determineSeverity(auditType, line)

	event := types.NewEvent("auditd", eventType, severity, line)
	event.SetHostname(hostname)

	fields := p.extractFields(line)

	if uid, ok := fields["uid"]; ok {
		event.User = uid
	} else if auid, ok := fields["auid"]; ok {
		event.User = auid
	}

	// comm/exe
	if comm, ok := fields["comm"]; ok {
		event.Process = strings.Trim(comm, `"`)
	} else if exe, ok := fields["exe"]; ok {
		event.Process = strings.Trim(exe, `"`)
	}

	if auditType == "EXECVE" {
		if a0, ok := fields["a0"]; ok {
			event.Command = strings.Trim(a0, `"`)
		}
	}

	return event, nil
}

func (p *AuditdParser) extractType(line string) string {
	matches := p.typeRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return "UNKNOWN"
}

func (p *AuditdParser) extractFields(line string) map[string]string {
	fields := make(map[string]string)
	matches := p.fieldRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 2 {
			fields[match[1]] = match[2]
		}
	}
	return fields
}

func (p *AuditdParser) mapAuditType(auditType string) string {
	switch auditType {
	case "SYSCALL":
		return "system_call"
	case "EXECVE":
		return "process_execution"
	case "USER_LOGIN":
		return "user_login"
	case "USER_LOGOUT":
		return "user_logout"
	case "USER_AUTH":
		return "user_authentication"
	case "USER_ACCT":
		return "user_account"
	case "CRED_ACQ":
		return "credential_acquisition"
	case "CRED_DISP":
		return "credential_disposal"
	case "USER_START":
		return "user_session_start"
	case "USER_END":
		return "user_session_end"
	case "USER_CMD":
		return "user_command"
	case "PATH":
		return "file_access"
	case "CWD":
		return "working_directory"
	case "PROCTITLE":
		return "process_title"
	default:
		return "audit_event"
	}
}

func (p *AuditdParser) determineSeverity(auditType, line string) string {
	highPriorityTypes := []string{
		"USER_LOGIN", "USER_AUTH", "CRED_ACQ", "CRED_DISP",
		"USER_CMD", "EXECVE",
	}

	for _, t := range highPriorityTypes {
		if auditType == t {
			if strings.Contains(line, "res=failed") ||
				strings.Contains(line, "success=no") {
				return "high"
			}
			return "medium"
		}
	}

	if auditType == "SYSCALL" {
		if strings.Contains(line, "syscall=59") || // execve
			strings.Contains(line, "syscall=322") || // execveat
			strings.Contains(line, "syscall=2") { // open
			return "medium"
		}
	}

	return "low"
}
