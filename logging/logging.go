package logging

import (
	"bufio"
	"bytes"
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

var (
	logFile   *os.File
	logLevel  string
	logger    *log.Logger
	preLogger *bytes.Buffer = new(bytes.Buffer) // Buffer temporaire pour les logs avant InitLogger
)

func InitLogger(logPath string, level string) error {
	logLevel = level

	// Préparer la destination des logs
	var writers []io.Writer
	writers = append(writers, os.Stdout) // Toujours stdout

	if logPath != "" {
		// Vérifier si `logPath` est un dossier
		info, err := os.Stat(logPath)
		if err == nil && info.IsDir() {
			// 📌 `logPath` est un dossier, créer un fichier dynamique dedans
			logPath = filepath.Join(logPath, "strigo_"+time.Now().Format("20060102_150405")+".log")
		}

		// Assurer que le dossier parent existe
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}

		// 📂 Ouvrir le fichier log
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", logPath, err)
		}
		writers = append(writers, logFile)
	}

	// Créer le logger avec tous les writers
	multiWriter := io.MultiWriter(writers...)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime)

	// ✅ 📌 **Correction : éviter l'affichage en double**
	if preLogger != nil {
		scanner := bufio.NewScanner(preLogger)
		for scanner.Scan() {
			line := scanner.Text()

			// 🔥 Filtrer les logs avant affichage
			if shouldLog(line) {
				if logFile != nil {
					logger.Println(line) // ✅ **Écrire seulement dans le fichier**
				} else {
					fmt.Println(line) // ✅ **Sinon, afficher une seule fois**
				}
			}
		}
		preLogger = nil // 🚀 On vide `preLogger` après traitement
	}

	LogDebug("[INFO] Logger initialized successfully.")
	return nil
}

// **shouldLog : Filtre les logs en fonction du `logLevel`**
func shouldLog(entry string) bool {
	if logLevel == DebugLevel {
		return true // Tout log est affiché en mode debug
	} else if logLevel == InfoLevel {
		return !strings.HasPrefix(entry, "[DEBUG]") // Ignore les logs DEBUG
	} else {
		return strings.HasPrefix(entry, "[ERROR]") // En mode error, ne log que les erreurs
	}
}
func PreLog(level string, format string, args ...interface{}) {
	if preLogger == nil {
		preLogger = new(bytes.Buffer) // S'assurer que le buffer existe
	}

	// 🔥 Filtrer selon le niveau de log configuré
	if (logLevel == InfoLevel && level == DebugLevel) || (logLevel == ErrorLevel && level != ErrorLevel) {
		return // ❌ Ignore DEBUG en INFO et tout sauf ERROR en ERROR
	}

	// ✅ Évite l'affichage brut du fichier TOML en filtrant le contenu
	logEntry := fmt.Sprintf("[%s] %s\n", level, fmt.Sprintf(format, args...))
	if !strings.HasPrefix(logEntry, "[DEBUG] 📜 Raw file content") {
		preLogger.WriteString(logEntry) // ✅ Ajout au buffer uniquement si pertinent
	}
}

// **LogError : Logue un message d'erreur**
func LogError(format string, v ...interface{}) {
	message := fmt.Sprintf("[ERROR] "+format, v...)
	if logger != nil {
		logger.Println(message)
	} else {
		PreLog("ERROR", format, v...)
	}
}

// **LogInfo : Logue un message d'information**
func LogInfo(format string, v ...interface{}) {
	message := fmt.Sprintf("[INFO] "+format, v...)
	if logger != nil {
		logger.Println(message)
	} else {
		PreLog("INFO", format, v...)
	}
}

// **LogDebug : Logue un message de debug**
func LogDebug(format string, v ...interface{}) {
	if logLevel == DebugLevel {
		message := fmt.Sprintf("[DEBUG] "+format, v...)
		if logger != nil {
			logger.Println(message)
		} else {
			PreLog("DEBUG", format, v...)
		}
	}
}

// **LogOutput : Affiche un message en console sans préfixe (ne passe pas par logger)**
func LogOutput(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(message)

	// Si le fichier de log est actif, écrire dedans aussi
	if logFile != nil {
		logFile.WriteString(message + "\n")
	}
}

// **SetPreLogLevel : Définit le niveau de log avant `InitLogger()`**
func SetPreLogLevel(level string) {
	logLevel = level
}
