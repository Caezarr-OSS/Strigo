package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strigo/config"
	"strigo/logging"
	"strings"
)

// SDKAsset represents an available version of an SDK
type SDKAsset struct {
	Version     string `json:"version"`
	DownloadUrl string `json:"downloadUrl"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
}

// NexusClient implements RepositoryClient for Nexus repositories
type NexusClient struct{}

// NexusAsset represents an asset returned by Nexus API
type NexusAsset struct {
	Path        string            `json:"path"`
	DownloadUrl string            `json:"downloadUrl"`
	Checksum    map[string]string `json:"checksum"`
}

// GetAvailableVersions fetches available versions of a JDK from a Nexus repository.
func (c *NexusClient) GetAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) ([]SDKAsset, error) {
	var sdkAssets []SDKAsset
	var ignoredFiles []string
	seenVersions := make(map[string]bool) // Pour suivre les versions déjà vues

	// Ensure apiURL is correctly formatted and replace placeholders
	apiURL := strings.ReplaceAll(registry.APIURL, "{repository}", repo.Repository)

	// Build final request URL
	requestURL := fmt.Sprintf("%s&path=%s", apiURL, repo.Path)

	logging.LogDebug("🔍 Querying Nexus API: %s", requestURL)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query Nexus API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nexus API returned %d: Check if the path %s exists in Nexus", resp.StatusCode, repo.Path)
	}

	// Parse JSON response
	var data struct {
		Items []NexusAsset `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	logging.LogDebug("🔍 Raw items from Nexus:")

	// Construire le chemin complet pour la distribution
	distributionPath := fmt.Sprintf("/%s/", repo.Path)

	for _, item := range data.Items {
		logging.LogDebug("   Path: %s", item.Path)

		// Vérifier si le chemin correspond à la distribution demandée
		if !strings.Contains(item.Path, distributionPath) {
			ignoredFiles = append(ignoredFiles, item.Path)
			continue
		}

		versionName := extractVersionName(item.Path)
		if versionName != "" {
			logging.LogDebug("   Extracted version: %s from path: %s", versionName, item.Path)
			// Vérifier si cette version a déjà été vue
			if !seenVersions[versionName] {
				seenVersions[versionName] = true
				sdkAsset := SDKAsset{
					Version:     versionName,
					DownloadUrl: item.DownloadUrl,
					Filename:    versionName,
					// Size sera ajouté plus tard si nécessaire
				}
				sdkAssets = append(sdkAssets, sdkAsset)
			}
		} else {
			ignoredFiles = append(ignoredFiles, item.Path)
		}
	}

	if len(ignoredFiles) > 0 {
		logging.LogDebug("❌ Ignored files:")
		for _, f := range ignoredFiles {
			logging.LogDebug("   - %s", f)
		}
	}

	// Filtrer les versions si un filtre est spécifié
	if versionFilter != "" {
		var filteredAssets []SDKAsset
		for _, asset := range sdkAssets {
			if strings.Contains(asset.Version, versionFilter) {
				filteredAssets = append(filteredAssets, asset)
			}
		}
		sdkAssets = filteredAssets
	}

	if len(sdkAssets) == 0 {
		if versionFilter != "" {
			return nil, fmt.Errorf("no version %s found for %s", versionFilter, repo.Path)
		}
		return nil, fmt.Errorf("no versions found for %s", repo.Path)
	}

	// Trier les versions
	sort.Slice(sdkAssets, func(i, j int) bool {
		return sdkAssets[i].Version > sdkAssets[j].Version
	})

	return sdkAssets, nil
}

// extractVersionName extracts the versioned filename from a Nexus path.
func extractVersionName(path string) string {
	// Handle different naming patterns
	patterns := []string{
		`corretto-(\d+\.\d+\.\d+\.\d+)`,             // For Corretto: 11.0.26.4.1
		`jdk-(\d+\.\d+\.\d+_\d+)`,                   // For Temurin: 11.0.26_4
		`jdk_x64_linux_hotspot_(\d+\.\d+\.\d+_\d+)`, // Alternative Temurin pattern
		`(\d+u\d+\w+)`,                              // For older versions: 8u442b06
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(path); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// FindAssetByVersion helps locate a specific asset in the version map
func FindAssetByVersion(versionMap map[string][]NexusAsset, targetVersion string) *NexusAsset {
	for _, assets := range versionMap {
		for _, asset := range assets {
			// Check if the file name contains the target version
			if strings.Contains(asset.Path, targetVersion) {
				return &asset
			}
		}
	}
	return nil
}
