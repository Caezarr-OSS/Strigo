package cmd

import (
	"os"
	"strigo/logging"
	"strigo/repository"

	"github.com/spf13/cobra"
)

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:   "available [type] <distribution> [version]",
	Short: "List available versions of a specific SDK",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if cfg == nil {
			logging.LogError("❌ Configuration is not loaded.")
			os.Exit(1)
		}

		sdkType := args[0]
		distribution := args[1]
		var versionFilter string
		if len(args) > 2 {
			versionFilter = args[2]
		}

		// Debugging log
		logging.LogDebug("🔍 Checking distribution: %s of type %s", distribution, sdkType)

		// Check if the distribution exists in the configuration
		sdkRepo, exists := cfg.SDKRepositories[distribution]
		if !exists {
			logging.LogError("❌ Distribution %s not found in configuration", distribution)
			return
		}

		// Get registry information
		registry, exists := cfg.Registries[sdkRepo.Registry]
		if !exists {
			logging.LogError("❌ Registry %s not found in configuration", sdkRepo.Registry)
			return
		}

		// ✅ Just call FetchAvailableVersions (it already logs versions)
		err := repository.FetchAvailableVersions(sdkRepo, registry, versionFilter)
		if err != nil {
			logging.LogError("❌ Error retrieving available versions: %v", err)
			return
		}
	},
}
