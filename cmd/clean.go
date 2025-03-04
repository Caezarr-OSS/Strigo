package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/config"
	"strigo/logging"
	"strings"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean invalid JAVA_HOME configuration",
	Long: `Clean invalid JAVA_HOME configuration. This command will:
1. Check if current JAVA_HOME points to a valid JDK installation
2. If not, remove JAVA_HOME from shell configuration
3. Inform user about the changes`,
	Run: clean,
}

func clean(cmd *cobra.Command, args []string) {
	if err := handleClean(); err != nil {
		ExitWithError(err)
	}
}

func handleClean() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Obtenir JAVA_HOME actuel
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		logging.LogInfo("ℹ️  No JAVA_HOME currently set")
		return nil
	}

	// Vérifier si le chemin existe
	if _, err := os.Stat(javaHome); os.IsNotExist(err) {
		logging.LogInfo("❌ Current JAVA_HOME points to non-existent path: %s", javaHome)
		return cleanJavaHome()
	}

	// Vérifier si c'est un JDK installé par strigo
	sdkInstallDir := cfg.General.SDKInstallDir
	if !strings.HasPrefix(javaHome, sdkInstallDir) {
		logging.LogInfo("⚠️  Current JAVA_HOME points to a JDK not managed by strigo: %s", javaHome)
		return nil
	}

	// Vérifier si le JDK est toujours valide
	relativePath := strings.TrimPrefix(javaHome, sdkInstallDir)
	parts := strings.Split(strings.Trim(relativePath, string(os.PathSeparator)), string(os.PathSeparator))

	if len(parts) < 3 {
		logging.LogInfo("❌ Invalid JDK path structure")
		return cleanJavaHome()
	}

	sdkType, distribution := parts[0], parts[1]

	// Vérifier si le type de SDK existe (accepter singulier ou pluriel)
	baseType := strings.TrimSuffix(sdkType, "s") // Enlever le 's' final si présent
	if _, exists := cfg.SDKTypes[baseType]; !exists {
		logging.LogInfo("❌ Invalid SDK type: %s", sdkType)
		return cleanJavaHome()
	}

	// Vérifier si la distribution existe
	if _, exists := cfg.SDKRepositories[distribution]; !exists {
		logging.LogInfo("❌ Invalid distribution: %s", distribution)
		return cleanJavaHome()
	}

	logging.LogInfo("✅ Current JAVA_HOME is valid: %s", javaHome)
	return nil
}

func cleanJavaHome() error {
	// Déterminer le shell de l'utilisateur
	shell := os.Getenv("SHELL")
	var rcFile string

	switch {
	case strings.HasSuffix(shell, "bash"):
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
	case strings.HasSuffix(shell, "zsh"):
		rcFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
	default:
		return fmt.Errorf("unsupported shell: %s. Please clean JAVA_HOME manually", shell)
	}

	// Lire le contenu actuel
	content, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read rc file: %w", err)
	}

	// Supprimer les lignes JAVA_HOME
	lines := strings.Split(string(content), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, "JAVA_HOME=") && !strings.Contains(line, "PATH=$JAVA_HOME") {
			newLines = append(newLines, line)
		}
	}

	// Écrire le nouveau contenu
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(rcFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update rc file: %w", err)
	}

	logging.LogInfo("✅ Successfully removed JAVA_HOME configuration")
	logging.LogInfo("ℹ️  Please run 'source %s' to apply the changes", rcFile)

	return nil
}
