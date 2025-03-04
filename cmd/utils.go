package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strigo/config"
	"strigo/logging"
)

// ListOutput structure pour la sortie JSON des commandes list et available
type ListOutput struct {
	Types         []string `json:"types,omitempty"`
	Distributions []string `json:"distributions,omitempty"`
	Versions      []string `json:"versions,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// outputJSON gère la sortie JSON pour toutes les commandes
func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// GetInstallPath retourne le chemin d'installation complet pour un SDK
func GetInstallPath(cfg *config.Config, sdkType, distribution, version string) (string, error) {
	// Vérifier si le type de SDK existe
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return "", fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Construire le chemin complet
	return filepath.Join(
		cfg.General.SDKInstallDir,
		sdkTypeConfig.InstallDir,
		distribution,
		version,
	), nil
}

// ExitWithError affiche l'erreur et quitte avec le code 1
func ExitWithError(err error) {
	if jsonOutput {
		outputJSON(ListOutput{Error: err.Error()})
	} else {
		logging.LogError("❌ %v", err)
	}
	os.Exit(1)
}
