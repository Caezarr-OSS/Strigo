package downloader

import (
	"fmt"
	"path/filepath"
	"strigo/downloader/cache"
	"strigo/downloader/core"
	"strigo/downloader/jdk"
	"strigo/downloader/network"
	"strigo/logging"
)

// Manager orchestre le processus de téléchargement et d'installation
type Manager struct {
	network     *network.Client
	extractor   *Extractor
	cache       *cache.Manager
	validator   *core.Validator
	certificates *jdk.CertificateManager
}

// NewManager crée une nouvelle instance de Manager
func NewManager() *Manager {
	return &Manager{
		network:     network.NewClient(),
		extractor:   NewExtractor(),
		cache:       cache.NewManager(),
		validator:   core.NewValidator(),
		certificates: jdk.NewCertificateManager(),
	}
}

// DownloadAndExtract gère le processus complet de téléchargement et d'installation
func (m *Manager) DownloadAndExtract(opts core.DownloadOptions) error {
	logging.LogDebug("🔍 Starting installation process for %s %s %s", opts.SDKType, opts.Distribution, opts.Version)

	// Vérifier la taille du fichier
	fileSize, err := m.network.GetFileSize(opts.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// Valider l'espace disponible
	if err := m.validator.ValidateSpace(fileSize, opts.CacheDir); err != nil {
		return fmt.Errorf("cache directory space check failed: %w", err)
	}
	if err := m.validator.ValidateSpace(fileSize, filepath.Dir(opts.InstallPath)); err != nil {
		return fmt.Errorf("install directory space check failed: %w", err)
	}

	// Préparer le cache
	cachePath, err := m.cache.PrepareCacheDirectory(opts.SDKType, opts.Distribution, opts.Version, opts.CacheDir)
	if err != nil {
		return fmt.Errorf("failed to prepare cache: %w", err)
	}

	// Télécharger le fichier
	cacheFile := filepath.Join(cachePath, filepath.Base(opts.DownloadURL))
	if err := m.network.DownloadFile(opts.DownloadURL, cacheFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Valider et créer le répertoire d'installation
	if err := m.validator.ValidateDirectories(opts.InstallPath); err != nil {
		return fmt.Errorf("failed to prepare installation directory: %w", err)
	}

	// Extraire l'archive
	if err := m.extractor.Extract(cacheFile, opts.InstallPath); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Nettoyer le cache si nécessaire
	if err := m.cache.CleanupCache(cachePath, opts.KeepCache); err != nil {
		logging.LogDebug("⚠️ Cache cleanup failed: %v", err)
	}

	// Configurer les certificats si nécessaire
	if opts.SDKType == "jdk" {
		if err := m.certificates.SetupCertificates(opts.InstallPath, opts.CertConfig); err != nil {
			logging.LogDebug("⚠️ Certificate setup failed: %v", err)
			logging.LogInfo("ℹ️ JDK installation is complete but certificates were not configured")
		}
	}

	logging.LogInfo("✅ Successfully installed %s %s version %s", opts.SDKType, opts.Distribution, opts.Version)
	logging.LogInfo("📂 Installation path: %s", opts.InstallPath)
	return nil
}
