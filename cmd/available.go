package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strigo/config"
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
		// Charger la configuration avant la validation
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

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

		output := &AvailableOutput{}

		// Si pas d'arguments, afficher les types de SDK disponibles
		if len(args) == 0 {
			return handleNoArgs(output)
		}

		sdkType := args[0]

		// Si seulement le type est fourni, afficher les distributions
		if len(args) == 1 {
			return handleTypeOnly(sdkType, output)
		}

		distribution := args[1]
		var versionFilter string
		if len(args) > 2 {
			versionFilter = args[2]
		}

		return handleFullCommand(sdkType, distribution, versionFilter, output)
	},
}

// Fonctions utilitaires
func getValidSDKTypes() []string {
	if cfg == nil {
		return []string{}
	}
	types := make([]string, 0, len(cfg.SDKTypes))
	for sdkType := range cfg.SDKTypes {
		types = append(types, sdkType)
	}
	sort.Strings(types)
	return types
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

// ExtractMajorVersion extrait la version majeure d'une chaîne de version
func ExtractMajorVersion(version string) string {
	logging.LogDebug("Extracting major version from: %s", version)

	// Si la version est vide, retourner vide
	if version == "" {
		logging.LogDebug("Empty version string")
		return ""
	}

	// Pour les versions Node.js qui commencent directement par un nombre (22.13.1)
	if firstDot := strings.Index(version, "."); firstDot != -1 {
		majorPart := version[:firstDot]
		if _, err := strconv.Atoi(majorPart); err == nil {
			logging.LogDebug("Found major version (direct number): %s", majorPart)
			return majorPart
		}
	}

	// Supprimer tout préfixe non numérique (comme "jdk-")
	version = strings.TrimLeft(version, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_")

	// Trouver le premier nombre qui représente la version majeure
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		// Pour les versions comme "8u442b06", extraire le 8
		majorPart := strings.Split(parts[0], "u")[0]
		if _, err := strconv.Atoi(majorPart); err == nil {
			logging.LogDebug("Found major version (after cleanup): %s", majorPart)
			return majorPart
		}
	}

	logging.LogDebug("No major version found")
	return ""
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
	versions, err := repository.FetchAvailableVersions(sdkRepo, registry, "", true)
	if err != nil {
		logging.LogError("❌ %v", err)
		return nil
	}

	logging.LogDebug("Found %d versions before filtering", len(versions))

	// Collecter toutes les versions majeures disponibles
	allMajorVersions := make(map[string]bool)
	for _, v := range versions {
		logging.LogDebug("Version before filtering: %s", v.Version)
		majorVersion := ExtractMajorVersion(v.Version)
		if majorVersion != "" {
			allMajorVersions[majorVersion] = true
		}
	}

	// Convertir en slice et trier
	var availableMajors []int
	for major := range allMajorVersions {
		if num, err := strconv.Atoi(major); err == nil {
			availableMajors = append(availableMajors, num)
		}
	}
	sort.Ints(availableMajors)

	// Filtrer les versions si un filtre est spécifié
	if versionFilter != "" {
		var filteredVersions []repository.SDKAsset
		for _, v := range versions {
			logging.LogDebug("Checking version %s against filter %s", v.Version, versionFilter)
			if ExtractMajorVersion(v.Version) == versionFilter {
				logging.LogDebug("  ✓ Version matches filter")
				filteredVersions = append(filteredVersions, v)
			} else {
				logging.LogDebug("  ✗ Version does not match filter")
			}
		}

		// Si aucune version ne correspond au filtre, afficher les versions disponibles
		if len(filteredVersions) == 0 {
			logging.LogOutput("❌ No version found matching major version %s", versionFilter)
			logging.LogOutput("")
			logging.LogOutput("💡 Available major versions are: %s", joinInts(availableMajors))
			return nil
		}

		versions = filteredVersions
		logging.LogDebug("Found %d versions after filtering", len(versions))
	}

	// Trier les versions
	sort.Slice(versions, func(i, j int) bool {
		return repository.CompareVersions(versions[i].Version, versions[j].Version)
	})

	output.Versions = versions

	displayVersions(versions, sdkType, distribution)
	return nil
}

// joinInts convertit une slice d'entiers en chaîne de caractères
func joinInts(numbers []int) string {
	var strNumbers []string
	for _, num := range numbers {
		strNumbers = append(strNumbers, strconv.Itoa(num))
	}
	return strings.Join(strNumbers, ", ")
}

func displayVersions(versions []repository.SDKAsset, sdkType, distribution string) {
	logging.LogDebug("Processing %d versions for display", len(versions))

	// Grouper les versions par version majeure
	versionGroups := make(map[string][]string)
	allMajorVersions := make(map[string]bool)

	// Récupérer toutes les versions majeures disponibles
	for _, asset := range versions {
		logging.LogDebug("Processing version: %s", asset.Version)
		majorVersion := ExtractMajorVersion(asset.Version)
		logging.LogDebug("  Extracted major version: %s", majorVersion)
		if majorVersion != "" {
			allMajorVersions[majorVersion] = true
			versionGroups[majorVersion] = append(versionGroups[majorVersion], asset.Version)
			logging.LogDebug("  Added to version groups. Current groups: %v", versionGroups)
		}
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

	// Si aucune version n'est trouvée
	if len(majorVersions) == 0 {
		logging.LogOutput("❌ No major version found matching your criteria")
		logging.LogOutput("")

		// Créer la liste des versions majeures disponibles
		var availableMajors []int
		for major := range allMajorVersions {
			if num, err := strconv.Atoi(major); err == nil {
				availableMajors = append(availableMajors, num)
			}
		}
		sort.Ints(availableMajors)

		// Convertir les versions en chaînes pour l'affichage
		var majorStrings []string
		for _, num := range availableMajors {
			majorStrings = append(majorStrings, strconv.Itoa(num))
		}

		logging.LogOutput("💡 Available major versions are: %s", strings.Join(majorStrings, ", "))
		return
	}

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
	logging.LogOutput(fmt.Sprintf("   strigo install %s %s [version]", sdkType, distribution))
}
