package cmd

import (
	"fmt"
	"os"
	"strigo/config"
	"strigo/logging"

	"github.com/spf13/cobra"
)

// Global config variable
var cfg *config.Config

// Root command
var rootCmd = &cobra.Command{
	Use:   "strigo",
	Short: "Strigo - SDK & JDK Version Manager",
}

func init() {
	// Pre-log important startup messages before logger is initialized
	logging.PreLog("DEBUG", "🔧 Initializing Strigo...")

	// Load configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Println("❌ Error loading configuration:", err)
		os.Exit(1)
	}

	// Initialize logger with config values
	err = logging.InitLogger(cfg.General.LogPath, cfg.General.LogLevel)
	if err != nil {
		fmt.Println("❌ Error initializing logger:", err)
		os.Exit(1)
	}

	// Ensure required directories exist
	err = config.EnsureDirectoriesExist(cfg)
	if err != nil {
		fmt.Println("❌ Error ensuring directories:", err)
		os.Exit(1)
	}

	// Add subcommands
	rootCmd.AddCommand(availableCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌ Error:", err)
		logging.LogError("❌ Error executing command: %v", err)
		os.Exit(1)
	}
}
