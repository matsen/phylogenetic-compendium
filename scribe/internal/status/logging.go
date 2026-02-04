package status

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultLogPath is the default path for action logs.
const DefaultLogPath = ".claude/authoring/logs/actions.jsonl"

// ActionLogger provides agent action logging.
type ActionLogger struct {
	path string
}

// NewActionLogger creates a new ActionLogger.
func NewActionLogger(path string) *ActionLogger {
	if path == "" {
		path = DefaultLogPath
	}
	return &ActionLogger{path: path}
}

// Log appends an action log entry.
func (l *ActionLogger) Log(taskID string, agentType AgentType, action, target string, result AgentActionResult, message *string) error {
	entry := AgentActionLog{
		LogID:     fmt.Sprintf("log-%d", time.Now().UnixNano()),
		TaskID:    taskID,
		AgentType: agentType,
		Action:    action,
		Target:    target,
		Result:    result,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Ensure directory exists
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	// Append to log file
	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write log entry: %w", err)
	}

	return nil
}

// ReadAll reads all log entries.
func (l *ActionLogger) ReadAll() ([]AgentActionLog, error) {
	file, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []AgentActionLog{}, nil
		}
		return nil, fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	var logs []AgentActionLog
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry AgentActionLog
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed entries
		}
		logs = append(logs, entry)
	}

	return logs, scanner.Err()
}

// ReadForTask reads log entries for a specific task.
func (l *ActionLogger) ReadForTask(taskID string) ([]AgentActionLog, error) {
	all, err := l.ReadAll()
	if err != nil {
		return nil, err
	}

	var filtered []AgentActionLog
	for _, entry := range all {
		if entry.TaskID == taskID {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

// Clear removes all log entries.
func (l *ActionLogger) Clear() error {
	err := os.Remove(l.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
