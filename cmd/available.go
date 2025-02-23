package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strigo/logging"
	"strigo/repository"
	"strings"

	"github.com/spf13/cobra"
)

// Structures pour la sortie JSON
type AvailableOutput struct {
	Types         []string              `json:"types,omitempty"`
	Distributions []string              `json:"distributions,omitempty"`
	Versions      []repository.SDKAsset `json:"versions,omitempty"`
	Error         string                `json:"error,omitempty"`
}

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:   "available [type] <distribution> [version]",
	Short: "List available versions of a specific SDK",
	Long: `List available versions of a specific SDK.
Examples:
  strigo available                  # List all available SDK types
  strigo available jdk             # List all available JDK distributions
  strigo available jdk temurin     # List all Temurin JDK versions
  strigo available jdk temurin 11  # List Temurin JDK versions containing "11"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 3 {
			return fmt.Errorf("too many arguments. Use 'strigo available --help' for usage")
		}

		if len(args) == 0 {
			return nil
		}

		// Validation du type de SDK
		sdkType := args[0]
		validTypes := getValidSDKTypes()
		if !contains(validTypes, sdkType) {
			return fmt.Errorf("invalid SDK type '%s'. Available types: %s", sdkType, strings.Join(validTypes, ", "))
		}

		if len(args) == 1 {
			return nil
		}

		// Validation de la distribution
		distribution := args[1]
		validDists := getValidDistributions(sdkType)
		if !contains(validDists, distribution) {
			return fmt.Errorf("invalid distribution '%s' for type '%s'. Available distributions: %s",
				distribution, sdkType, strings.Join(validDists, ", "))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			return fmt.Errorf("configuration is not loaded")
		}

		output := AvailableOutput{}

		// Si pas d'arguments, afficher les types de SDK disponibles
		if len(args) == 0 {
			return handleNoArgs(&output)
		}

		sdkType := args[0]

		// Si seulement le type est fourni, afficher les distributions
		if len(args) == 1 {
			return handleTypeOnly(sdkType, &output)
		}

		distribution := args[1]
		var versionFilter string
		if len(args) > 2 {
			versionFilter = args[2]
		}

		return handleFullCommand(sdkType, distribution, versionFilter, &output)
	},
}

// Fonctions utilitaires
func getValidSDKTypes() []string {
	if cfg == nil {
		return []string{}
	}
	types := make(map[string]bool)
	for _, repo := range cfg.SDKRepositories {
		if repo.Type != "" {
			types[repo.Type] = true
		}
	}
	return mapToSortedSlice(types)
}

func getValidDistributions(sdkType string) []string {
	if cfg == nil {
		return []string{}
	}
	var dists []string
	for name, repo := range cfg.SDKRepositories {
		if repo.Type == sdkType {
			dists = append(dists, name)
		}
	}
	sort.Strings(dists)
	return dists
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func mapToSortedSlice(m map[string]bool) []string {
	var slice []string
	for k := range m {
		slice = append(slice, k)
	}
	sort.Strings(slice)
	return slice
}

// Handlers pour chaque cas d'utilisation
func handleNoArgs(output *AvailableOutput) error {
	types := getValidSDKTypes()
	output.Types = types

	if len(types) > 0 {
		logging.LogOutput("Available SDK types:")
		logging.LogOutput("─────────────────────")
		for _, sdkType := range types {
			logging.LogOutput("✅ %s", sdkType)
		}
		logging.LogOutput("")
	} else {
		logging.LogOutput("No SDK types available")
	}
	return nil
}

func handleTypeOnly(sdkType string, output *AvailableOutput) error {
	for name, repo := range cfg.SDKRepositories {
		if repo.Type == sdkType {
			output.Distributions = append(output.Distributions, name)
		}
	}

	if len(output.Distributions) > 0 {
		logging.LogOutput("Available %s distributions:", sdkType)
		logging.LogOutput("─────────────────────────")
		for _, dist := range output.Distributions {
			logging.LogOutput("✅ %s", dist)
		}
	}
	return nil
}

func handleFullCommand(sdkType, distribution, versionFilter string, output *AvailableOutput) error {
	// Check if the distribution exists
	sdkRepo, exists := cfg.SDKRepositories[distribution]
	if !exists {
		err := fmt.Errorf("distribution %s not found in configuration", distribution)
		logging.LogError("❌ %v", err)
		return nil
	}

	// Get registry information
	registry, exists := cfg.Registries[sdkRepo.Registry]
	if !exists {
		err := fmt.Errorf("registry %s not found in configuration", sdkRepo.Registry)
		logging.LogError("❌ %v", err)
		return nil
	}

	// Fetch available versions
	versions, err := repository.FetchAvailableVersions(sdkRepo, registry, versionFilter, true)
	if err != nil {
		logging.LogError("❌ %v", err)
		return nil
	}

	// Trier les versions
	sort.Slice(versions, func(i, j int) bool {
		return repository.CompareVersions(versions[i].Version, versions[j].Version)
	})

	output.Versions = versions

	displayVersions(versions)
	return nil
}

func displayVersions(versions []repository.SDKAsset) {
	// Grouper les versions par version majeure
	versionGroups := make(map[string][]string)
	for _, asset := range versions {
		majorVersion := repository.ExtractMajorVersion(asset.Version)
		versionGroups[majorVersion] = append(versionGroups[majorVersion], asset.Version)
	}

	// Obtenir les versions majeures triées
	var majorVersions []int
	for major := range versionGroups {
		if num, err := strconv.Atoi(major); err == nil {
			majorVersions = append(majorVersions, num)
		}
	}
	sort.Ints(majorVersions)

	logging.LogOutput("🔹 Available versions:")
	logging.LogOutput("─────────────────────────")

	// Afficher les versions par groupe
	for _, majorNum := range majorVersions {
		major := strconv.Itoa(majorNum)
		versions := versionGroups[major]

		// Trier les versions dans chaque groupe
		sort.Slice(versions, func(i, j int) bool {
			return repository.CompareVersions(versions[i], versions[j])
		})

		logging.LogOutput("-%d :", majorNum)
		for _, version := range versions {
			logging.LogOutput("    ✅ %s", version)
		}
		logging.LogOutput("") // Ligne vide entre les groupes
	}

	logging.LogOutput("💡 To install a specific version:")
	logging.LogOutput("   strigo install jdk [distribution] [version]")
}
