package downloader

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strigo/logging"
	"strings"
)

// DownloadAndExtract handles the complete process of downloading and extracting an SDK
func DownloadAndExtract(downloadURL, cacheDir, installPath string, sdkType, distribution, version string, keepCache bool, certConfig CertConfig) error {
	// Get file size before downloading
	fileSize, err := getFileSize(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// Check available space in both cache and install directories
	if err := checkDiskSpace(fileSize, cacheDir); err != nil {
		return fmt.Errorf("insufficient space in cache directory: %w", err)
	}
	if err := checkDiskSpace(fileSize, filepath.Dir(installPath)); err != nil {
		return fmt.Errorf("insufficient space in installation directory: %w", err)
	}

	// Create cache path that mirrors installation structure
	cachePath := filepath.Join(cacheDir, sdkType, distribution, version)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache file path with full structure
	cacheFile := filepath.Join(cachePath, filepath.Base(downloadURL))

	// Check if already in cache
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		logging.LogInfo("🚀 Downloading SDK...")
		if err := downloadFile(downloadURL, cacheFile); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
	} else {
		logging.LogInfo("📦 Using cached version from %s", cacheFile)
	}

	// Create installation directory
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Extract the archive
	logging.LogInfo("📂 Extracting SDK...")
	if err := extractTarGz(cacheFile, installPath); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Après l'extraction réussie
	if !keepCache {
		logging.LogInfo("🧹 Cleaning up cache file...")
		if err := cleanupCacheDirectory(cachePath); err != nil {
			logging.LogDebug("Failed to clean cache: %v", err)
		}
	}

	// Si symlink est activé, gérer les certificats
	if certConfig.Enabled {
		if err := setupJDKCertificates(installPath, certConfig); err != nil {
			return fmt.Errorf("failed to setup certificates: %w", err)
		}
	}

	return nil
}

// downloadFile downloads a file from URL to a local path
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz extracts a .tar.gz file to a destination directory
func extractTarGz(tarPath, destPath string) error {
	if !filepath.IsAbs(destPath) {
		return fmt.Errorf("destination path must be absolute")
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Get the target path
		target := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destPath)) {
				return fmt.Errorf("invalid tar path: %s", header.Name)
			}

			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}

	return nil
}

type CertConfig struct {
	Enabled           bool
	JDKSecurityPath   string
	SystemCacertsPath string
}

func setupJDKCertificates(installPath string, config CertConfig) error {
	if config.JDKSecurityPath == "" {
		return fmt.Errorf("jdk_security_path is not configured")
	}
	if config.SystemCacertsPath == "" {
		return fmt.Errorf("system_cacerts_path is not configured")
	}

	// Find the actual JDK directory (it's usually in a subdirectory after extraction)
	entries, err := os.ReadDir(installPath)
	if err != nil {
		return fmt.Errorf("failed to read JDK directory: %w", err)
	}

	// Look for the JDK directory (usually starts with "jdk")
	var jdkDir string
	for _, entry := range entries {
		if entry.IsDir() && (strings.HasPrefix(entry.Name(), "jdk") || strings.HasPrefix(entry.Name(), "java")) {
			jdkDir = entry.Name()
			break
		}
	}

	if jdkDir == "" {
		return fmt.Errorf("could not find JDK directory in %s", installPath)
	}

	// Build the complete path to the security directory
	certPath := filepath.Join(installPath, jdkDir, config.JDKSecurityPath)

	logging.LogDebug("🔐 Setting up certificates:")
	logging.LogDebug("   - JDK directory: %s", jdkDir)
	logging.LogDebug("   - Certificate path: %s", certPath)
	logging.LogDebug("   - System certificates: %s", config.SystemCacertsPath)

	// Verify system certificates exist
	if _, err := os.Stat(config.SystemCacertsPath); os.IsNotExist(err) {
		return fmt.Errorf("system certificates not found: %s", config.SystemCacertsPath)
	}

	// Remove existing certificates if they exist
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing certificates: %w", err)
	}

	// Create symlink
	if err := os.Symlink(config.SystemCacertsPath, certPath); err != nil {
		return fmt.Errorf("failed to create certificate symlink: %w", err)
	}

	logging.LogInfo("✅ Certificates successfully configured")
	return nil
}

func getFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get file size: HTTP %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}

// Download downloads a file from a URL and saves it to the specified path
func Download(url, destPath string) error {
	// Get file size
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("unable to get file information: %w", err)
	}
	defer resp.Body.Close()

	fileSize := int64(0)
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		fileSize, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			logging.LogDebug("⚠️ Unable to parse Content-Length, will proceed without disk space check")
		}
	}

	// Check available disk space only if we know the file size
	if fileSize > 0 {
		if err := checkDiskSpace(fileSize, filepath.Dir(destPath)); err != nil {
			return err
		}
	}

	// ... existing code ...

	return nil
}

// Nouvelle fonction pour nettoyer le cache et les répertoires parents vides
func cleanupCacheDirectory(cachePath string) error {
	// Supprimer le répertoire de cache et son contenu
	if err := os.RemoveAll(cachePath); err != nil {
		return fmt.Errorf("failed to remove cache directory: %w", err)
	}

	// Remonter l'arborescence et supprimer les répertoires vides
	currentPath := filepath.Dir(cachePath) // Remonte d'un niveau
	for {
		// Vérifier si le répertoire est vide
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			break // En cas d'erreur, on arrête
		}

		if len(entries) > 0 {
			break // Si le répertoire n'est pas vide, on arrête
		}

		// Supprimer le répertoire vide
		if err := os.Remove(currentPath); err != nil {
			break // En cas d'erreur, on arrête
		}

		// Remonter d'un niveau
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break // Si on est à la racine, on arrête
		}
		currentPath = parentPath
	}

	return nil
}
