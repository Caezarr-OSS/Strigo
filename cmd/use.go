package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/logging"
	"strings"

	"github.com/spf13/cobra"
)

var (
	setEnvVar bool
	unsetEnv  bool
)

func init() {
	useCmd.Flags().BoolVarP(&setEnvVar, "set-env", "e", false, "Set environment variables in shell configuration file (~/.bashrc or ~/.zshrc)")
	useCmd.Flags().BoolVar(&unsetEnv, "unset", false, "Remove environment variables from shell configuration file")
}

var useCmd = &cobra.Command{
	Use:   "use [type] [distribution] [version]",
	Short: "Set a specific SDK version as active",
	Long: `Set a specific SDK version as active. For example:
strigo use jdk temurin 11.0.24_8

This will create a symbolic link to the specified version.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if unsetEnv {
			if len(args) != 1 || (args[0] != "jdk" && args[0] != "node") {
				return fmt.Errorf("\n❌ Invalid arguments for --unset\n\n" +
					"Usage:\n" +
					"  strigo use [jdk|node] --unset")
			}
			return nil
		}

		if len(args) != 3 {
			return fmt.Errorf("\n❌ Invalid number of arguments\n\n" +
				"Usage:\n" +
				"  strigo use [type] [distribution] [version]\n\n" +
				"Example:\n" +
				"  strigo use jdk temurin 11.0.24_8\n\n" +
				"To see installed versions:\n" +
				"  strigo list jdk temurin")
		}
		return nil
	},
	Run: use,
	Example: `  # Use Temurin JDK 11
  strigo use jdk temurin 11.0.24_8

  # Use Corretto JDK 8
  strigo use jdk corretto 8u442b06`,
}

func use(cmd *cobra.Command, args []string) {
	if unsetEnv {
		if err := handleUnset(args[0]); err != nil {
			ExitWithError(err)
		}
		return
	}

	if err := handleUse(args[0], args[1], args[2]); err != nil {
		ExitWithError(err)
	}
}

func getSDKBinPath(basePath string, sdkType string) (string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// SDK directory selection logic
	var sdkDir string
	dirCount := 0
	
	// Count directories and remember the first one
	for _, entry := range entries {
		if entry.IsDir() {
			dirCount++
			// If it's the first directory, remember it
			if sdkDir == "" {
				sdkDir = entry.Name()
			}
		}
	}
	
	// If multiple directories exist, it's ambiguous
	if dirCount > 1 {
		sdkDir = ""
	}

	if sdkDir == "" {
		return "", fmt.Errorf("could not find %s directory in %s", strings.ToUpper(sdkType), basePath)
	}

	return filepath.Join(basePath, sdkDir), nil
}

func findRcFile() (string, error) {
	// Check if shell_config_path is set in config
	if cfg.General.ShellConfigPath != "" {
		return cfg.General.ShellConfigPath, nil
	}

	// Auto-detect based on current shell
	shell := os.Getenv("SHELL")
	home := os.Getenv("HOME")

	// List of possible RC files
	var rcFiles []string

	// Determine the order based on the shell
	if strings.HasSuffix(shell, "zsh") {
		rcFiles = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bashrc"), // fallback
		}
	} else if strings.HasSuffix(shell, "bash") {
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"), // fallback
		}
	} else {
		// Unrecognized shell, try both
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
		}
	}

	// Find the first existing RC file
	for _, file := range rcFiles {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("no shell configuration file found (.zshrc or .bashrc). Please set shell_config_path in strigo.toml")
}

func handleUnset(sdkType string) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not loaded")
	}

	if sdkType != "jdk" && sdkType != "node" {
		return fmt.Errorf("unset is only supported for JDK and Node.js")
	}

	rcFile, err := findRcFile()
	if err != nil {
		return fmt.Errorf("could not find shell configuration file: %w", err)
	}

	// Expand tilde if present
	expandedPath := rcFile
	if strings.HasPrefix(rcFile, "~") {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		expandedPath = filepath.Join(home, rcFile[1:])
	}

	// Read the current content
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", expandedPath, err)
	}

	// Remove the Strigo configuration block
	lines := strings.Split(string(content), "\n")
	var newLines []string
	var removed bool
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// If we find the Strigo comment
		if strings.Contains(line, fmt.Sprintf("# Added by Strigo - %s configuration", strings.ToUpper(sdkType))) {
			// Skip this line and the next two
			i += 2 // +2 because the loop will do +1
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !removed {
		logging.LogInfo("ℹ️  No Strigo %s configuration found in %s", strings.ToUpper(sdkType), rcFile)
		return nil
	}

	// Write the file
	newContent := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(expandedPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update %s: %w", expandedPath, err)
	}

	logging.LogInfo("✅ Successfully removed Strigo %s configuration from %s", strings.ToUpper(sdkType), expandedPath)
	logging.LogInfo("ℹ️  To apply these changes, run: source %s", expandedPath)

	return nil
}

func handleUse(sdkType, distribution, version string) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not loaded")
	}

	// Check if the SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Build the installation path
	installPath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir, distribution, version)

	// Check if the SDK is installed
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("version %s %s %s is not installed", sdkType, distribution, version)
	}

	// Get the binary path
	sdkPath, err := getSDKBinPath(installPath, sdkType)
	if err != nil {
		return fmt.Errorf("failed to find SDK binary path: %w", err)
	}

	// Create the symbolic link
	linkPath := filepath.Join(cfg.General.SDKInstallDir, fmt.Sprintf("current-%s", sdkType))

	// Remove the existing link if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing symbolic link: %w", err)
		}
	}

	// Create the new link
	if err := os.Symlink(sdkPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symbolic link: %w", err)
	}

	logging.LogInfo("✅ Successfully set %s %s version %s as active", sdkType, distribution, version)

	// If --set-env is specified, configure the environment variables
	if setEnvVar {
		if err := configureEnvironment(sdkType, sdkPath); err != nil {
			return fmt.Errorf("failed to configure environment: %w", err)
		}
	} else {
		if sdkType == "jdk" {
			logging.LogInfo("ℹ️  To use this JDK, set these environment variables:")
			logging.LogInfo("   export JAVA_HOME=%s", sdkPath)
			logging.LogInfo("   export PATH=$JAVA_HOME/bin:$PATH")
			logging.LogInfo("")
			logging.LogInfo("💡 Or use --set-env to set them automatically in your shell configuration")
		} else if sdkType == "node" {
			logging.LogInfo("ℹ️  To use this Node.js version, set these environment variables:")
			logging.LogInfo("   export NODE_HOME=%s", sdkPath)
			logging.LogInfo("   export PATH=$NODE_HOME/bin:$PATH")
			logging.LogInfo("")
			logging.LogInfo("💡 Or use --set-env to set them automatically in your shell configuration")
		}
	}

	return nil
}

func configureEnvironment(sdkType, sdkPath string) error {
	// Find the appropriate RC file
	rcFile, err := findRcFile()
	if err != nil {
		return err
	}

	// Expand tilde if present
	expandedPath := rcFile
	if strings.HasPrefix(rcFile, "~") {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		expandedPath = filepath.Join(home, rcFile[1:])
	}

	// Read the current content
	content, err := os.ReadFile(expandedPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read rc file: %w", err)
	}

	// Prepare the new lines
	var envVar string
	if sdkType == "jdk" {
		envVar = "JAVA_HOME"
	} else if sdkType == "node" {
		envVar = "NODE_HOME"
	}

	newConfig := fmt.Sprintf("\n# Added by Strigo - %s configuration\nexport %s=%s\nexport PATH=$%s/bin:$PATH\n",
		strings.ToUpper(sdkType), envVar, sdkPath, envVar)

	// Remove the old configuration if it exists
	lines := strings.Split(string(content), "\n")
	var newLines []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, fmt.Sprintf("# Added by Strigo - %s configuration", strings.ToUpper(sdkType))) {
			i += 2 // Skip next two lines
			continue
		}
		newLines = append(newLines, line)
	}

	// Add the new configuration
	newContent := strings.Join(newLines, "\n") + newConfig

	// Write the new content
	if err := os.WriteFile(expandedPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update rc file: %w", err)
	}

	logging.LogInfo("✅ Successfully configured environment in %s", expandedPath)
	logging.LogInfo("ℹ️  To apply these changes, run: source %s", expandedPath)

	return nil
}
