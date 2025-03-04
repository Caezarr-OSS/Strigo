package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/downloader"
	"strigo/downloader/core"
	"strigo/logging"
	"strigo/repository"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [type] [distribution] [version]",
	Short: "Install a specific SDK version",
	Long: `Install a specific SDK version. For example:
	strigo install jdk temurin 11.0.24_8
	strigo install jdk corretto 8u442b06

Available SDK types:
	jdk     Java Development Kit

Available distributions for jdk:
	temurin    Eclipse Temurin (AdoptOpenJDK)
	corretto   Amazon Corretto`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("\n❌ Invalid number of arguments\n\n" +
				"Usage:\n" +
				"  strigo install [type] [distribution] [version]\n\n" +
				"Example:\n" +
				"  strigo install jdk temurin 11.0.24_8\n\n" +
				"To see available versions:\n" +
				"  strigo available jdk temurin")
		}
		return nil
	},
	Run: install,
	Example: `  # Install Temurin JDK 11
  strigo install jdk temurin 11.0.24_8

  # Install Corretto JDK 8
  strigo install jdk corretto 8u442b06

  # To see available versions:
  strigo available jdk temurin`,
}

func install(cmd *cobra.Command, args []string) {
	sdkType := args[0]
	distribution := args[1]
	version := args[2]

	if err := handleInstall(sdkType, distribution, version); err != nil {
		logging.LogError("❌ Error executing command: %v", err)
		return
	}
}

func handleInstall(sdkType, distribution, version string) error {
	if cfg == nil {
		logging.LogError("❌ Configuration is not loaded")
		return nil
	}

	logging.LogDebug("🔧 Starting installation of %s %s version %s", sdkType, distribution, version)

	// Check if the SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		logging.LogError("❌ SDK type %s not found in configuration", sdkType)
		return nil
	}

	// Check if the distribution exists
	sdkRepo, exists := cfg.SDKRepositories[distribution]
	if !exists {
		logging.LogError("❌ Distribution %s not found in configuration", distribution)
		return nil
	}

	// Verify that the distribution's type matches the requested type
	if sdkRepo.Type != sdkTypeConfig.Type {
		logging.LogError("❌ Distribution %s is not of type %s", distribution, sdkType)
		return nil
	}

	// Get registry information
	registry, exists := cfg.Registries[sdkRepo.Registry]
	if !exists {
		logging.LogError("❌ Registry %s not found in configuration", sdkRepo.Registry)
		return nil
	}

	// Fetch available versions with filter
	assets, err := repository.FetchAvailableVersions(sdkRepo, registry, version, true) // true pour supprimer l'affichage
	if err != nil {
		logging.LogError("❌ Failed to fetch versions: %v", err)
		return nil
	}

	// Find exact version match
	var matchedAsset *repository.SDKAsset
	for _, asset := range assets {
		if asset.Version == version {
			matchedAsset = &asset
			break
		}
	}

	if matchedAsset == nil {
		logging.LogError("❌ Version %s not found", version)
		logging.LogInfo("💡 Use 'strigo available %s %s' to see available versions", sdkType, distribution)
		return nil
	}

	logging.LogInfo("✅ Found version %s, preparing for installation...", version)

	// Get installation path
	installPath, err := GetInstallPath(cfg, sdkType, distribution, version)
	if err != nil {
		logging.LogError("❌ Failed to get installation path: %v", err)
		return nil
	}

	// Check if already installed
	if _, err := os.Stat(installPath); err == nil {
		logging.LogError("❌ Version %s is already installed at %s", version, installPath)
		return nil
	}

	// Create installation directory
	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		logging.LogError("❌ Failed to create installation directory: %v", err)
		return nil
	}

	// Prepare certificate configuration
	certConfig := core.CertConfig{
		JDKSecurityPath:   cfg.General.JDKSecurityPath,
		SystemCacertsPath: cfg.General.SystemCacertsPath,
	}

	// Download and extract
	manager := downloader.NewManager()
	opts := core.DownloadOptions{
		DownloadURL:  matchedAsset.DownloadUrl,
		CacheDir:     cfg.General.CacheDir,
		InstallPath:  installPath,
		SDKType:      sdkType,
		Distribution: distribution,
		Version:      version,
		KeepCache:    cfg.General.KeepCache,
		CertConfig:   certConfig,
	}
	err = manager.DownloadAndExtract(opts)

	if err != nil {
		logging.LogError("❌ Installation failed: %v", err)
		// Cleanup on failure
		os.RemoveAll(installPath)
		return nil
	}

	// Pour les JDKs, gérer les certificats
	if sdkType == "jdk" {
		// Trouver le dossier JDK extrait
		entries, err := os.ReadDir(installPath)
		if err != nil {
			return fmt.Errorf("failed to read installation directory: %w", err)
		}

		var jdkDir string
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "jdk") {
				jdkDir = entry.Name()
				break
			}
		}

		if jdkDir == "" {
			return fmt.Errorf("could not find JDK directory in %s", installPath)
		}

		// Utiliser le chemin complet pour les certificats
		jdkPath := filepath.Join(installPath, jdkDir)
		jdkSecPath := filepath.Join(jdkPath, cfg.General.JDKSecurityPath)

		// 1. Supprimer les certificats par défaut
		logging.LogDebug("🗑️ Removing default JDK certificates...")
		if err := os.RemoveAll(jdkSecPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove default certificates: %w", err)
		}

		// 2. Créer le lien symbolique vers les certificats système
		logging.LogDebug("🔗 Creating link to system certificates...")
		if err := os.MkdirAll(filepath.Dir(jdkSecPath), 0755); err != nil {
			return fmt.Errorf("failed to create security directory: %w", err)
		}

		if err := os.Symlink(cfg.General.SystemCacertsPath, jdkSecPath); err != nil {
			return fmt.Errorf("failed to create symlink to system certificates: %w", err)
		}
		logging.LogInfo("✅ Successfully linked system certificates")
	}

	logging.LogInfo("✅ Successfully installed %s %s version %s", sdkType, distribution, version)
	logging.LogInfo("📂 Installation path: %s", installPath)
	logging.LogInfo("ℹ️  To set this version as active, run: strigo use %s %s %s", sdkType, distribution, version)

	return nil
}
