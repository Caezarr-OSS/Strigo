package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strigo/logging"

	"github.com/spf13/cobra"
)

var (
	cleanCache bool
)

var removeCmd = &cobra.Command{
	Use:   "remove [tool] [vendor] [version]",
	Short: "Remove a specific version of a tool",
	Long: `Remove a specific version of a tool. For example:
strigo remove jdk temurin 11.0.26_4`,
	Args: cobra.ExactArgs(3),
	Run:  remove,
}

func init() {
	removeCmd.Flags().BoolVar(&cleanCache, "clean-cache", false, "Also clean cache directory for the removed version")
}

func remove(cmd *cobra.Command, args []string) {
	tool := args[0]
	vendor := args[1]
	version := args[2]

	logging.LogDebug("🗑️ Attempting to remove %s %s version %s", tool, vendor, version)

	if err := handleRemove(tool, vendor, version); err != nil {
		logging.LogError("Failed to remove version: %v", err)
		return
	}

	logging.LogInfo("✅ Successfully removed %s %s version %s", tool, vendor, version)
}

func handleRemove(sdkType, distribution, version string) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not loaded")
	}

	// Vérifier si le type de SDK existe
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Construire le chemin d'installation
	installPath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir, distribution, version)
	logging.LogDebug("🔍 Checking installation path: %s", installPath)

	// Vérifier si le dossier existe
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		logging.LogDebug("❌ Installation path not found: %s", installPath)

		// Vérifier si c'est peut-être le dossier décompressé
		extractedPath := filepath.Join(installPath, fmt.Sprintf("jdk-%s", version))
		logging.LogDebug("🔍 Checking extracted path: %s", extractedPath)

		// Lister le contenu du dossier parent pour debug
		parentDir := filepath.Dir(installPath)
		if entries, err := os.ReadDir(parentDir); err == nil {
			logging.LogDebug("📂 Content of %s:", parentDir)
			for _, entry := range entries {
				logging.LogDebug("   - %s", entry.Name())
			}
		}

		logging.LogError("Failed to remove version: version %s %s %s is not installed", sdkType, distribution, version)
		logging.LogDebug("❌ Neither installation path nor extracted path exists")
		return fmt.Errorf("version not installed")
	}

	logging.LogDebug("🗑️ Removing SDK from: %s", installPath)

	// Supprimer le dossier
	if err := os.RemoveAll(installPath); err != nil {
		logging.LogError("❌ Failed to remove SDK: %v", err)
		logging.LogDebug("Error details: %v", err)
		return err
	}

	// Clean cache if requested
	if cleanCache {
		cachePath := filepath.Join(cfg.General.CacheDir, sdkType, distribution, version)
		if _, err := os.Stat(cachePath); err == nil {
			logging.LogDebug("Cleaning up cache directory: %s", cachePath)
			if err := os.RemoveAll(cachePath); err != nil {
				logging.LogDebug("Failed to remove cache directory: %v", err)
			}
		}
	}

	// Check if vendor directory is empty
	vendorPath := filepath.Join(cfg.General.SDKInstallDir, sdkType, distribution)
	if isEmpty, _ := isDirEmpty(vendorPath); isEmpty {
		logging.LogDebug("Removing empty vendor directory: %s", vendorPath)
		os.Remove(vendorPath)
	}

	// Check if tool directory is empty
	toolPath := filepath.Join(cfg.General.SDKInstallDir, sdkType)
	if isEmpty, _ := isDirEmpty(toolPath); isEmpty {
		logging.LogDebug("Removing empty tool directory: %s", toolPath)
		os.Remove(toolPath)
	}

	return nil
}

func isActiveVersion(tool, vendor, version string) bool {
	// TODO: This function will be implemented with the 'use' command
	// For now, we assume no version is active
	return false
}

func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil // Directory is not empty
	}
	if err == io.EOF {
		return true, nil // Directory is empty
	}
	return false, err
}
