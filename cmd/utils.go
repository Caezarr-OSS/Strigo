package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strigo/config"
	"strigo/logging"
)

// ListOutput structure for JSON output of list and available commands
type ListOutput struct {
	Types         []string `json:"types,omitempty"`
	Distributions []string `json:"distributions,omitempty"`
	Versions      []string `json:"versions,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// outputJSON handles JSON output for all commands
func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// GetInstallPath returns the complete installation path for an SDK
func GetInstallPath(cfg *config.Config, sdkType, distribution, version string) (string, error) {
	// Check if SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return "", fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Build complete path
	return filepath.Join(
		cfg.General.SDKInstallDir,
		sdkTypeConfig.InstallDir,
		distribution,
		version,
	), nil
}

// ExitWithError displays the error and exits with code 1
func ExitWithError(err error) {
	if jsonOutput {
		if outputErr := outputJSON(ListOutput{Error: err.Error()}); outputErr != nil {
			logging.LogError("Error outputting JSON: %v", outputErr)
		}
	} else {
		logging.LogError("❌ %v", err)
	}
	os.Exit(1)
}
