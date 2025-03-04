package cmd

import (
	"encoding/json"
	"fmt"
)

// Variables globales pour le mode JSON
var (
	jsonOutput bool // Flag pour la sortie en JSON
	jsonLogs   bool // Flag pour les logs en JSON
)

// GetJsonOutput retourne la valeur du flag JSON
func GetJsonOutput() bool {
	return jsonOutput
}

// CommandOutput structure pour la sortie JSON
type CommandOutput struct {
	Types         []string `json:"types,omitempty"`
	Distributions []string `json:"distributions,omitempty"`
	Versions      []string `json:"versions,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// OutputJSON gère la sortie JSON pour toutes les commandes
func OutputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}
