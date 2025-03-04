package logging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Log levels
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	ErrorLevel = "error"
)

// LogEntry représente une entrée de log au format JSON
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Component string      `json:"component,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

var (
	logFile   *os.File
	logLevel  string
	logger    *log.Logger
	preLogger *bytes.Buffer = new(bytes.Buffer)
	useJSON   bool // Indique si on utilise le format JSON
)

func InitLogger(logPath string, level string, jsonFormat bool) error {
	logLevel = level
	useJSON = jsonFormat

	// Préparer la destination des logs
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if logPath != "" {
		info, err := os.Stat(logPath)
		if err == nil && info.IsDir() {
			logPath = filepath.Join(logPath, "strigo_"+time.Now().Format("20060102_150405")+".log")
		}

		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}

		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", logPath, err)
		}
		writers = append(writers, logFile)
	}

	multiWriter := io.MultiWriter(writers...)
	logger = log.New(multiWriter, "", 0) // Pas de préfixe car on gère nous-mêmes le format

	if preLogger != nil {
		scanner := bufio.NewScanner(preLogger)
		for scanner.Scan() {
			line := scanner.Text()
			if shouldLog(line) {
				if logFile != nil {
					logger.Println(line)
				} else {
					fmt.Println(line)
				}
			}
		}
		preLogger = nil
	}

	LogDebug("[INFO] Logger initialized successfully.")
	return nil
}

func shouldLog(entry string) bool {
	if logLevel == DebugLevel {
		return true
	} else if logLevel == InfoLevel {
		return !strings.HasPrefix(entry, "[DEBUG]")
	} else {
		return strings.HasPrefix(entry, "[ERROR]")
	}
}

func writeLog(level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	
	if useJSON {
		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     level,
			Message:   message,
		}
		
		if jsonData, err := json.Marshal(entry); err == nil {
			if logger != nil {
				logger.Println(string(jsonData))
			} else {
				PreLog(level, string(jsonData))
			}
		}
	} else {
		formattedMessage := fmt.Sprintf("[%s] %s", level, message)
		if logger != nil {
			logger.Println(formattedMessage)
		} else {
			PreLog(level, message)
		}
	}
}

func LogError(format string, v ...interface{}) {
	writeLog("ERROR", format, v...)
}

func LogInfo(format string, v ...interface{}) {
	writeLog("INFO", format, v...)
}

func LogDebug(format string, v ...interface{}) {
	if logLevel == DebugLevel {
		writeLog("DEBUG", format, v...)
	}
}

// LogOutputWithData affiche un message avec des données structurées optionnelles
func LogOutputWithData(format string, data interface{}, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	
	if useJSON {
		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "OUTPUT",
			Message:   message,
		}
		if data != nil {
			entry.Data = data
		}
		
		if jsonData, err := json.Marshal(entry); err == nil {
			fmt.Println(string(jsonData))
			if logFile != nil {
				logFile.WriteString(string(jsonData) + "\n")
			}
		}
	} else {
		fmt.Println(message)
		if logFile != nil {
			logFile.WriteString(message + "\n")
		}
	}
}

// LogOutput est maintenant un wrapper autour de LogOutputWithData sans données
func LogOutput(format string, args ...interface{}) {
	LogOutputWithData(format, nil, args...)
}

func PreLog(level string, format string, args ...interface{}) {
	if preLogger == nil {
		preLogger = new(bytes.Buffer)
	}

	if (logLevel == InfoLevel && level == DebugLevel) || (logLevel == ErrorLevel && level != ErrorLevel) {
		return
	}

	var logEntry string
	if useJSON {
		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     level,
			Message:   fmt.Sprintf(format, args...),
		}
		if jsonData, err := json.Marshal(entry); err == nil {
			logEntry = string(jsonData) + "\n"
		}
	} else {
		logEntry = fmt.Sprintf("[%s] %s\n", level, fmt.Sprintf(format, args...))
	}

	if !strings.HasPrefix(logEntry, "[DEBUG] 📜 Raw file content") {
		preLogger.WriteString(logEntry)
	}
}

func SetPreLogLevel(level string) {
	logLevel = level
}
