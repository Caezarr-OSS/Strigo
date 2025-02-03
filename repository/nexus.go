package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strigo/config"
	"strigo/logging"
	"strings"
)

// NexusClient implements RepositoryClient for Nexus repositories
type NexusClient struct{}

// GetAvailableVersions fetches available versions of a JDK from a Nexus repository.
func (c *NexusClient) GetAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) ([]string, error) {
	versionMap := make(map[string][]string)
	var allVersions []string
	var ignoredFiles []string

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
		Items []struct {
			Path string `json:"path"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	// Normalize repo.Path to match Nexus paths
	repoPath := repo.Path
	if !strings.HasPrefix(repoPath, "/") {
		repoPath = "/" + repoPath
	}

	// Process retrieved versions
	for _, item := range data.Items {
		logging.LogDebug("🔹 Raw path from Nexus: %s", item.Path)

		// Ensure the file belongs to the correct directory
		if !strings.HasPrefix(item.Path, repoPath) {
			ignoredFiles = append(ignoredFiles, item.Path)
			continue
		}

		// Extract version name
		versionName := extractVersionName(item.Path)
		if versionName != "" {
			// Extract major version (8, 11, 17, etc.)
			versionMajor := extractMajorVersion(versionName)
			versionMap[versionMajor] = append(versionMap[versionMajor], versionName)
			allVersions = append(allVersions, versionName) // Store versions for return
		}
	}

	// Log ignored files for debugging
	if len(ignoredFiles) > 0 {
		logging.LogDebug("❌ Ignored files (not matching path %s):", repoPath)
		for _, f := range ignoredFiles {
			logging.LogDebug("   - %s", f)
		}
	}

	// Final check and structured output
	if len(versionMap) == 0 {
		logging.LogError("❌ No versions found for %s in Nexus", repo.Path)
		return nil, fmt.Errorf("no versions found for %s", repo.Path)
	}

	// Sort major versions numerically before displaying them
	availableMajors := []int{}
	for major := range versionMap {
		majorNum, err := strconv.Atoi(major)
		if err == nil {
			availableMajors = append(availableMajors, majorNum)
		}
	}
	sort.Ints(availableMajors) // Sort numerically

	// Convert back to string for display
	sortedMajorVersions := []string{}
	for _, major := range availableMajors {
		sortedMajorVersions = append(sortedMajorVersions, strconv.Itoa(major))
	}

	// If a version filter is applied but not found, list available versions
	if versionFilter != "" {
		if versions, exists := versionMap[versionFilter]; exists {
			return versions, nil
		}
		logging.LogError("❌ No version %s found for %s", versionFilter, repo.Path)
		logging.LogInfo("🔹 Available major versions: %s", strings.Join(sortedMajorVersions, ", "))
		return nil, fmt.Errorf("no version %s found for %s. Available versions: %s", versionFilter, repo.Path, strings.Join(sortedMajorVersions, ", "))
	}

	// ✅ Don't log here! Logging should be done at the call site.
	return allVersions, nil
}

// extractVersionName extracts the versioned filename from a Nexus path.
func extractVersionName(path string) string {
	segments := strings.Split(path, "/")
	if len(segments) > 0 {
		return segments[len(segments)-1] // Return the last element (filename)
	}
	return ""
}

// extractMajorVersion extracts the major version (8, 11, 17, etc.) from a versioned filename.
func extractMajorVersion(versionName string) string {
	re := regexp.MustCompile(`(\d+)(\.\d+)*`)
	match := re.FindStringSubmatch(versionName)
	if len(match) > 1 {
		return match[1] // Return the first matched number (major version)
	}
	return "unknown"
}
