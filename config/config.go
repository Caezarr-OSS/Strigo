package config

import (
	"fmt"
	"os"
	"strigo/logging"

	"github.com/pelletier/go-toml"
)

// **GeneralConfig : Structure des paramètres généraux**
type GeneralConfig struct {
	LogLevel      string `toml:"log_level"`
	SDKInstallDir string `toml:"sdk_install_dir"`
	CacheDir      string `toml:"cache_dir"`
	LogPath       string `toml:"log_path"`
}

// **Config : Structure représentant `strigo.toml`**
type Config struct {
	General         GeneralConfig            `toml:"general"`
	Registries      map[string]Registry      `toml:"registries"`
	SDKRepositories map[string]SDKRepository `toml:"sdk_repositories"`
}

// **Registry : Structure d'un registre distant**
type Registry struct {
	Type   string `toml:"type"`
	APIURL string `toml:"api_url"`
}

// **SDKRepository : Structure d'un SDK référencé**
type SDKRepository struct {
	Type       string `toml:"type"`
	Registry   string `toml:"registry"`
	Repository string `toml:"repository"`
	Path       string `toml:"path"`
	Symlink    bool   `toml:"symlink"`
}

// **LoadConfig : Charge et parse le fichier `strigo.toml`**
func LoadConfig() (*Config, error) {
	// 🔍 Détecter le fichier de configuration
	configPath := os.Getenv("STRIGO_CONFIG_PATH")
	if configPath == "" {
		configPath = "strigo.toml" // Par défaut
	}

	// 🚀 Prelog pour capture avant InitLogger
	logging.PreLog("DEBUG", "📂 Loading configuration from: %s", configPath)

	// 📄 Lire le fichier de configuration
	file, err := os.ReadFile(configPath)
	if err != nil {
		logging.PreLog("ERROR", "❌ Failed to read config file: %v", err)
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 📜 Debug: Afficher le contenu brut du fichier
	logging.PreLog("DEBUG", "📜 Raw file content:\n%s", string(file))

	// 🔄 Désérialiser le fichier TOML
	var cfg Config
	err = toml.Unmarshal(file, &cfg)
	if err != nil {
		logging.PreLog("ERROR", "❌ Failed to parse config file: %v", err)
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 🔍 Debug: Afficher la structure décodée
	logging.PreLog("DEBUG", "🔍 Decoded Config: %+v", cfg)

	// 🛑 Vérification des champs obligatoires (⚠️ `LogPath` peut être vide)
	if cfg.General.SDKInstallDir == "" || cfg.General.CacheDir == "" {
		logging.PreLog("ERROR", "❌ Configuration values are empty! Check your `strigo.toml`.")
		logging.PreLog("DEBUG", "🔍 Debug: SDKInstallDir=%q | CacheDir=%q", cfg.General.SDKInstallDir, cfg.General.CacheDir)
		return nil, fmt.Errorf("one or more required configuration paths are empty")
	}

	// ✅ Appliquer le log level temporaire pour filtrer `PreLog()`
	logging.SetPreLogLevel(cfg.General.LogLevel)

	logging.PreLog("DEBUG", "✅ Configuration successfully loaded.")
	return &cfg, nil
}

// **EnsureDirectoriesExist : Vérifie et crée les répertoires nécessaires**
func EnsureDirectoriesExist(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("❌ configuration is nil, cannot ensure directories")
	}

	// 🔍 Debug: Affichage des chemins des dossiers
	logging.LogDebug("🔍 Checking directory paths: LogPath=%s, SDKInstallDir=%s, CacheDir=%s",
		cfg.General.LogPath, cfg.General.SDKInstallDir, cfg.General.CacheDir)

	// 🛑 Vérifier que les chemins obligatoires ne sont pas vides
	if cfg.General.SDKInstallDir == "" || cfg.General.CacheDir == "" {
		return fmt.Errorf("❌ one or more required directory paths are empty in configuration")
	}

	// 📂 Liste des répertoires obligatoires à créer
	mandatoryPaths := []string{cfg.General.SDKInstallDir, cfg.General.CacheDir}

	// 📌 `LogPath` est facultatif, on l'ajoute seulement s'il est défini
	if cfg.General.LogPath != "" {
		mandatoryPaths = append(mandatoryPaths, cfg.General.LogPath)
	} else {
		logging.LogDebug("⚠️ LogPath is empty. Logs will be written only to stdout.")
	}

	// 🔄 Création des dossiers si nécessaire
	for _, path := range mandatoryPaths {
		logging.LogDebug("📂 Ensuring directory exists: %s", path)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			logging.LogError("❌ Failed to create directory %s: %v", path, err)
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	logging.LogDebug("✅ Directories verified.")
	return nil
}
