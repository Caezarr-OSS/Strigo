package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strigo/config"
	"strigo/logging"
	"strigo/repository"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [type] [distribution]",
	Short: "List installed SDK versions",
	Long: `List installed SDK versions. For example:
strigo list              # List all installed SDKs
strigo list jdk         # List all installed JDK distributions
strigo list jdk temurin # List installed Temurin JDK versions`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 2 {
			return fmt.Errorf("\n❌ Too many arguments\n\n" +
				"Usage:\n" +
				"  strigo list                    # List all SDK types\n" +
				"  strigo list jdk               # List all JDK distributions\n" +
				"  strigo list jdk temurin      # List Temurin JDK versions\n")
		}
		return nil
	},
	Run: list,
	Example: `  # List all SDK types
  strigo list

  # List all JDK distributions
  strigo list jdk

  # List installed Temurin JDK versions
  strigo list jdk temurin`,
}

func list(cmd *cobra.Command, args []string) {
	if err := handleList(args); err != nil {
		ExitWithError(err)
	}
}

func handleList(args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	output := ListOutput{}

	switch len(args) {
	case 0:
		return listSDKTypes(cfg, &output)
	case 1:
		return listDistributions(cfg, args[0], &output)
	case 2:
		return listVersions(cfg, args[0], args[1], &output)
	default:
		return fmt.Errorf("too many arguments")
	}
}

func listSDKTypes(cfg *config.Config, output *ListOutput) error {
	var types []string
	for sdkType := range cfg.SDKTypes {
		types = append(types, sdkType)
	}
	sort.Strings(types)
	output.Types = types

	if jsonOutput {
		return outputJSON(output)
	}

	if len(types) == 0 {
		logging.LogInfo("No SDKs installed")
		return nil
	}

	logging.LogInfo("Available SDK types:")
	logging.LogInfo("─────────────────────")
	for _, sdkType := range types {
		logging.LogInfo("✅ %s", sdkType)
	}
	logging.LogInfo("")

	return nil
}

func listDistributions(cfg *config.Config, sdkType string, output *ListOutput) error {
	// Vérifier si le type de SDK existe
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Construire le chemin de base
	basePath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir)

	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			if jsonOutput {
				return outputJSON(output)
			}
			logging.LogInfo("No distributions installed for %s", sdkType)
			return nil
		}
		return fmt.Errorf("failed to read distributions directory: %w", err)
	}

	var dists []string
	for _, entry := range entries {
		if entry.IsDir() {
			dists = append(dists, entry.Name())
		}
	}
	sort.Strings(dists)
	output.Distributions = dists

	if len(dists) == 0 {
		logging.LogInfo("No distributions installed for %s", sdkType)
		return nil
	}

	logging.LogInfo("Installed %s distributions:", sdkType)
	logging.LogInfo("─────────────────────────────")
	for _, dist := range dists {
		logging.LogInfo("✅ %s", dist)
	}
	logging.LogInfo("")

	return nil
}

func listVersions(cfg *config.Config, sdkType, distribution string, output *ListOutput) error {
	// Vérifier si le type de SDK existe
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Construire le chemin de base
	basePath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir, distribution)

	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			logging.LogInfo("No versions installed for %s %s", sdkType, distribution)
			return nil
		}
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	// Trier les versions
	sort.Slice(versions, func(i, j int) bool {
		return repository.CompareVersions(versions[i], versions[j])
	})
	output.Versions = versions

	if len(versions) == 0 {
		logging.LogInfo("No versions installed for %s %s", sdkType, distribution)
		return nil
	}

	// Grouper les versions par version majeure
	versionGroups := make(map[string][]string)
	for _, version := range versions {
		majorVersion := repository.ExtractMajorVersion(version)
		versionGroups[majorVersion] = append(versionGroups[majorVersion], version)
	}

	// Obtenir les versions majeures triées
	var majorVersions []int
	for major := range versionGroups {
		if num, err := strconv.Atoi(major); err == nil {
			majorVersions = append(majorVersions, num)
		}
	}
	sort.Ints(majorVersions)

	logging.LogInfo("Installed %s %s versions:", sdkType, distribution)
	logging.LogInfo("─────────────────────────────────")

	for _, majorNum := range majorVersions {
		major := strconv.Itoa(majorNum)
		versions := versionGroups[major]

		// Trier les versions dans chaque groupe
		sort.Slice(versions, func(i, j int) bool {
			return repository.CompareVersions(versions[i], versions[j])
		})

		logging.LogInfo("-%s :", major)
		for _, version := range versions {
			logging.LogInfo("    ✅ %s", version)
		}
		logging.LogInfo("")
	}

	return nil
}
