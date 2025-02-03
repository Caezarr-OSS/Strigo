package repository

import (
	"fmt"
	"sort"
	"strigo/config"
	"strigo/logging"
)

// RepositoryClient defines the interface for fetching available versions
type RepositoryClient interface {
	GetAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) ([]string, error)
}

// FetchAvailableVersions selects the appropriate repository client and applies filtering if necessary
func FetchAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) error {
	var client RepositoryClient

	switch registry.Type {
	case "nexus":
		client = &NexusClient{}
	default:
		logging.LogError("❌ Unsupported repository type: %s", registry.Type)
		return fmt.Errorf("unsupported repository type: %s", registry.Type)
	}

	// Retrieve versions from Nexus
	allVersions, err := client.GetAvailableVersions(repo, registry, versionFilter)
	if err != nil {
		return err
	}

	// ✅ Group and display versions properly
	if len(allVersions) == 0 {
		logging.LogError("❌ No versions found for %s", repo.Path)
		return fmt.Errorf("no versions found for %s", repo.Path)
	}

	// Group versions by major version
	groupedVersions := make(map[string][]string)
	for _, version := range allVersions {
		major := extractMajorVersion(version)
		groupedVersions[major] = append(groupedVersions[major], version)
	}

	// Sort major versions
	sortedMajors := []string{}
	for major := range groupedVersions {
		sortedMajors = append(sortedMajors, major)
	}
	sort.Strings(sortedMajors)

	// Log the grouped versions
	logging.LogOutput("🔹 Available versions:")
	for _, major := range sortedMajors {
		sort.Strings(groupedVersions[major]) // Sort versions within each major group
		logging.LogOutput("  - %s:", major)
		for _, v := range groupedVersions[major] {
			logging.LogOutput("    ✅ %s", v)
		}
	}

	return nil
}
