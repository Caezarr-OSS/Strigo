package core

import (
	"fmt"
	"os"
)

// Validator gère les validations système
type Validator struct{}

// NewValidator crée une nouvelle instance de Validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateSpace vérifie l'espace disponible
func (v *Validator) ValidateSpace(fileSize int64, directory string) error {
	return CheckDiskSpace(fileSize, directory)
}

// ValidateDirectories vérifie et crée les répertoires nécessaires
func (v *Validator) ValidateDirectories(installPath string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}
	return nil
}
