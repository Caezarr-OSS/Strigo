package cmd

import (
	"fmt"
	"strigo/config"
	"strigo/logging"

	"github.com/spf13/cobra"
)

// Global config variable
var cfg *config.Config

// Root command
var rootCmd = &cobra.Command{
	Use:           "strigo",
	Short:         "Strigo - SDK & JDK Version Manager",
	Long:          `Strigo is a command-line tool that helps you manage multiple versions of different SDKs (like JDK) on your system.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Ensure required directories exist
		if err := config.EnsureDirectoriesExist(cfg); err != nil {
			return fmt.Errorf("error ensuring directories: %w", err)
		}

		// Initialize logger with JSON format if requested
		if err := logging.InitLogger(cfg.General.LogPath, cfg.General.LogLevel, jsonOutput || jsonLogs); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		return nil
	},
}

func init() {
	// Pre-log important startup messages before logger is initialized
	logging.PreLog("DEBUG", "Initializing Strigo...")

	// Add subcommands
	rootCmd.AddCommand(availableCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(listCmd)

	// Allow flags to be placed after arguments
	rootCmd.Flags().SetInterspersed(true)

	// Add flags
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&jsonLogs, "json-logs", false, "Output logs in JSON format")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ExitWithError(err)
	}
}
