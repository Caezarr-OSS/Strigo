package repository

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strigo/config"
	"strigo/logging"
	"strings"
)

// RepositoryClient defines the interface for fetching available versions
type RepositoryClient interface {
	GetAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) ([]SDKAsset, error)
}

// FetchAvailableVersions fetches available versions with optional JSON output control
func FetchAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string, opts ...bool) ([]SDKAsset, error) {
	var client RepositoryClient

	// Par défaut, on affiche les versions (jsonOutput = false)
	jsonOutput := false
	if len(opts) > 0 {
		jsonOutput = opts[0]
	}

	switch registry.Type {
	case "nexus":
		client = &NexusClient{}
	default:
		logging.LogError("❌ Unsupported repository type: %s", registry.Type)
		return nil, fmt.Errorf("unsupported repository type: %s", registry.Type)
	}

	assets, err := client.GetAvailableVersions(repo, registry, versionFilter)
	if err != nil {
		return nil, err
	}

	// Si on n'est pas en mode JSON, on affiche les versions
	if !jsonOutput {
		displayVersions(assets)
	}

	return assets, nil
}

// displayVersions handles the user-friendly output
func displayVersions(assets []SDKAsset) {
	// Créer une map pour regrouper par version majeure
	versionGroups := make(map[string][]string)

	// Extraire la version majeure et regrouper
	for _, asset := range assets {
		majorVersion := extractMajorVersion(asset.Version)
		versionGroups[majorVersion] = append(versionGroups[majorVersion], asset.Version)
	}

	// Obtenir les versions majeures triées numériquement
	var majorVersions []int
	for major := range versionGroups {
		if num, err := strconv.Atoi(major); err == nil {
			majorVersions = append(majorVersions, num)
		}
	}
	sort.Ints(majorVersions)

	logging.LogOutput("🔹 Available versions:")
	for _, majorNum := range majorVersions {
		major := strconv.Itoa(majorNum)
		versions := versionGroups[major]

		// Trier les versions dans chaque groupe
		sort.Slice(versions, func(i, j int) bool {
			return CompareVersions(versions[i], versions[j])
		})

		logging.LogOutput("  - %s:", major)
		for _, version := range versions {
			logging.LogOutput("    ✅ %s", version)
		}
	}

	logging.LogOutput("\n💡 To install a specific version:")
	logging.LogOutput("   strigo install jdk [distribution] [version]")
}

// ExtractMajorVersion extrait la version majeure d'une version complète
func ExtractMajorVersion(version string) string {
	patterns := []string{
		`^(\d+)\..*`, // Pour 11.0.26_4, 21.0.6_7
		`^(\d+)u.*`,  // Pour 8u442b06
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(version); len(matches) > 1 {
			return matches[1]
		}
	}
	return "unknown"
}

// Garder une version privée pour une utilisation interne
func extractMajorVersion(version string) string {
	return ExtractMajorVersion(version)
}

// CompareVersions compare deux versions et retourne true si v1 est plus ancienne que v2
func CompareVersions(v1, v2 string) bool {
	// Normaliser les versions pour gérer les différents formats
	v1Parts := strings.Split(strings.Replace(strings.Replace(v1, "u", ".", -1), "_", ".", -1), ".")
	v2Parts := strings.Split(strings.Replace(strings.Replace(v2, "u", ".", -1), "_", ".", -1), ".")

	// Comparer chaque partie numérique
	minLen := len(v1Parts)
	if len(v2Parts) < minLen {
		minLen = len(v2Parts)
	}

	for i := 0; i < minLen; i++ {
		n1, err1 := strconv.Atoi(v1Parts[i])
		n2, err2 := strconv.Atoi(v2Parts[i])

		// Si une partie n'est pas un nombre, passer à la suivante
		if err1 != nil || err2 != nil {
			continue
		}

		if n1 != n2 {
			return n1 < n2 // Ordre croissant
		}
	}

	// Si toutes les parties sont égales, la version avec moins de parties est considérée plus ancienne
	return len(v1Parts) < len(v2Parts)
}
