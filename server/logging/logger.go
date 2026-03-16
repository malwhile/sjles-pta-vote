package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	logLevel  = INFO
	logPrefix = ""
)

func levelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Init initializes the logger with a log level from environment variable
// LOG_LEVEL can be: DEBUG, INFO, WARN, ERROR (default: INFO)
func Init() {
	levelEnv := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch levelEnv {
	case "DEBUG":
		logLevel = DEBUG
	case "WARN":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	case "INFO", "":
		logLevel = INFO
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(0) // We'll format our own timestamps
}

func shouldLog(level LogLevel) bool {
	return level >= logLevel
}

func formatLog(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05Z07:00")
	return fmt.Sprintf("%s | %s | %s", timestamp, levelString(level), message)
}

// Debug logs a debug-level message
func Debug(message string) {
	if shouldLog(DEBUG) {
		log.Println(formatLog(DEBUG, message))
	}
}

// Debugf logs a formatted debug-level message
func Debugf(format string, args ...interface{}) {
	if shouldLog(DEBUG) {
		log.Println(formatLog(DEBUG, fmt.Sprintf(format, args...)))
	}
}

// Info logs an info-level message
func Info(message string) {
	if shouldLog(INFO) {
		log.Println(formatLog(INFO, message))
	}
}

// Infof logs a formatted info-level message
func Infof(format string, args ...interface{}) {
	if shouldLog(INFO) {
		log.Println(formatLog(INFO, fmt.Sprintf(format, args...)))
	}
}

// Warn logs a warning-level message
func Warn(message string) {
	if shouldLog(WARN) {
		log.Println(formatLog(WARN, message))
	}
}

// Warnf logs a formatted warning-level message
func Warnf(format string, args ...interface{}) {
	if shouldLog(WARN) {
		log.Println(formatLog(WARN, fmt.Sprintf(format, args...)))
	}
}

// Error logs an error-level message
func Error(message string) {
	if shouldLog(ERROR) {
		log.Println(formatLog(ERROR, message))
	}
}

// Errorf logs a formatted error-level message
func Errorf(format string, args ...interface{}) {
	if shouldLog(ERROR) {
		log.Println(formatLog(ERROR, fmt.Sprintf(format, args...)))
	}
}

// Audit logs an audit trail event with structured information
// Example: Audit("CREATE_POLL", "admin@example.com", "Question: Budget?", true)
func Audit(action, user, details string, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	message := fmt.Sprintf("AUDIT | %s | user=%s | %s | %s", action, user, details, status)
	Info(message)
}
