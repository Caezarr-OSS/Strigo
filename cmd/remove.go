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

	// Check if SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Build installation path
	installPath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir, distribution, version)
	logging.LogDebug("🔍 Checking installation path: %s", installPath)

	// Check if directory exists
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		logging.LogDebug("❌ Installation path not found: %s", installPath)

		// Check if it might be the extracted directory
		extractedPath := filepath.Join(installPath, fmt.Sprintf("jdk-%s", version))
		logging.LogDebug("🔍 Checking extracted path: %s", extractedPath)

		// List parent directory content for debug
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

	// Remove directory
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

func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil // Directory is empty
	}
	return false, err // Either error or directory not empty
}
