// Package logging provides structured logging for MemoNexus.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel represents a log level.
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// Logger provides structured JSON logging.
type Logger struct {
	mu       sync.Mutex
	out      io.Writer
	minLevel LogLevel
}

var (
	// global logger instance
	global *Logger
	once   sync.Once
)

// Init initializes the global logger.
func Init(out io.Writer, minLevel LogLevel) {
	once.Do(func() {
		global = &Logger{
			out:      out,
			minLevel: minLevel,
		}
	})
}

// Get returns the global logger instance.
func Get() *Logger {
	if global == nil {
		Init(os.Stdout, LevelInfo)
	}
	return global
}

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp string  `json:"timestamp"`
	Level    string  `json:"level"`
	Message  string  `json:"message"`
	Error    string  `json:"error,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// log writes a log entry at the specified level.
func (l *Logger) log(level LogLevel, message string, err error, context map[string]interface{}) {
	// Check minimum level
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:    string(level),
		Message:  message,
		Context:  context,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Marshal to JSON
	data, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		log.Printf("Failed to marshal log entry: %v\n", jsonErr)
		return
	}

	// Write output
	fmt.Fprintln(l.out, string(data))
}

// shouldLog checks if a level should be logged.
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	return levels[level] >= levels[l.minLevel]
}

// Debug logs a debug message.
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	ctx := l.getContext(context...)
	l.log(LevelDebug, message, nil, ctx)
}

// Info logs an info message.
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	ctx := l.getContext(context...)
	l.log(LevelInfo, message, nil, ctx)
}

// Warn logs a warning message.
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	ctx := l.getContext(context...)
	l.log(LevelWarn, message, nil, ctx)
}

// Error logs an error message.
func (l *Logger) Error(message string, err error, context ...map[string]interface{}) {
	ctx := l.getContext(context...)
	l.log(LevelError, message, err, ctx)
}

// getContext merges multiple context maps.
func (l *Logger) getContext(context ...map[string]interface{}) map[string]interface{} {
	if len(context) == 0 {
		return nil
	}
	if len(context) == 1 {
		return context[0]
	}
	// Merge all contexts
	merged := make(map[string]interface{})
	for _, c := range context {
		for k, v := range c {
			merged[k] = v
		}
	}
	return merged
}

// Convenience functions using global logger

func Debug(message string, context ...map[string]interface{}) {
	Get().Debug(message, context...)
}

func Info(message string, context ...map[string]interface{}) {
	Get().Info(message, context...)
}

func Warn(message string, context ...map[string]interface{}) {
	Get().Warn(message, context...)
}

func Error(message string, err error, context ...map[string]interface{}) {
	Get().Error(message, err, context...)
}
