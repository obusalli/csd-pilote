package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
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

func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// LogEntry represents a log entry in the database
type LogEntry struct {
	ID        uint      `gorm:"primaryKey"`
	Level     string    `gorm:"type:varchar(10);not null;index"`
	Message   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null;index"`
}

func (LogEntry) TableName() string {
	return "logs"
}

type Logger struct {
	level       Level
	outputs     []io.Writer
	gormDB      *gorm.DB
	useDatabase bool
}

func NewLogger(level Level, outputs []io.Writer, gormDB *gorm.DB, useDatabase bool) *Logger {
	return &Logger{
		level:       level,
		outputs:     outputs,
		gormDB:      gormDB,
		useDatabase: useDatabase,
	}
}

func (l *Logger) log(level Level, message string) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	// Write to all configured outputs (console, file)
	for _, output := range l.outputs {
		output.Write([]byte(logMessage))
	}

	// Write to database if enabled
	if l.useDatabase && l.gormDB != nil {
		go l.logToDatabase(level, message)
	}
}

func (l *Logger) logToDatabase(level Level, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := LogEntry{
		Level:     level.String(),
		Message:   message,
		CreatedAt: time.Now(),
	}

	if err := l.gormDB.WithContext(ctx).Create(&logEntry).Error; err != nil {
		// Use standard log here to avoid infinite recursion
		log.Printf("Failed to write log to database: %v", err)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if len(args) == 0 {
		l.log(DEBUG, format)
	} else {
		l.log(DEBUG, fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if len(args) == 0 {
		l.log(INFO, format)
	} else {
		l.log(INFO, fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if len(args) == 0 {
		l.log(WARN, format)
	} else {
		l.log(WARN, fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if len(args) == 0 {
		l.log(ERROR, format)
	} else {
		l.log(ERROR, fmt.Sprintf(format, args...))
	}
}

// CreateLogFile creates and returns a file writer for logging
func CreateLogFile(logPath string, maxSizeMB int) (*os.File, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Check file size and rotate if necessary
	info, err := file.Stat()
	if err == nil && info.Size() > int64(maxSizeMB*1024*1024) {
		file.Close()
		rotateLog(logPath)
		file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file after rotation: %w", err)
		}
	}

	return file, nil
}

func rotateLog(logPath string) {
	timestamp := time.Now().Format("20060102-150405")
	newPath := fmt.Sprintf("%s.%s", logPath, timestamp)
	os.Rename(logPath, newPath)
}

// InitializeLogTable creates the logs table if it doesn't exist AutoMigrate
func InitializeLogTable(ctx context.Context, gormDB *gorm.DB) error {
	if err := gormDB.WithContext(ctx).AutoMigrate(&LogEntry{}); err != nil {
		return fmt.Errorf("failed to create logs table: %w", err)
	}
	return nil
}

// Global logger instance
var globalLogger *Logger

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(l *Logger) {
	globalLogger = l
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// Return a default logger if not initialized
		return NewLogger(INFO, []io.Writer{os.Stdout}, nil, false)
	}
	return globalLogger
}

// Global logging functions for convenience
func Info(format string, args ...interface{}) {
	GetGlobalLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetGlobalLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetGlobalLogger().Error(format, args...)
}

func Debug(format string, args ...interface{}) {
	GetGlobalLogger().Debug(format, args...)
}
