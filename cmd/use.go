package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/config"
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

	var sdkDir string
	for _, entry := range entries {
		if entry.IsDir() {
			if (sdkType == "jdk" && strings.HasPrefix(entry.Name(), "jdk")) ||
				(sdkType == "node" && strings.HasPrefix(entry.Name(), "node")) {
				sdkDir = entry.Name()
				break
			}
		}
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

	// Liste des fichiers RC possibles
	var rcFiles []string

	// Déterminer l'ordre en fonction du shell
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
		// Shell non reconnu, essayer les deux
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
		}
	}

	// Chercher le premier fichier RC existant
	for _, file := range rcFiles {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("no shell configuration file found (.zshrc or .bashrc). Please set shell_config_path in strigo.toml")
}

func handleUnset(sdkType string) error {
	// Load configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	if sdkType != "jdk" && sdkType != "node" {
		return fmt.Errorf("unset is only supported for JDK and Node.js")
	}

	rcFile, err := findRcFile()
	if err != nil {
		return fmt.Errorf("could not find shell configuration file: %w", err)
	}

	// Expand tilde if present
	expandedPath, err := config.ExpandTilde(rcFile)
	if err != nil {
		return fmt.Errorf("failed to expand path %s: %w", rcFile, err)
	}

	// Lire le contenu actuel
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", expandedPath, err)
	}

	// Supprimer le bloc de configuration Strigo
	lines := strings.Split(string(content), "\n")
	var newLines []string
	var removed bool
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// Si on trouve le commentaire Strigo
		if strings.Contains(line, fmt.Sprintf("# Added by Strigo - %s configuration", strings.ToUpper(sdkType))) {
			// On saute cette ligne et les 2 suivantes
			i += 2 // +2 car la boucle fera +1
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !removed {
		logging.LogInfo("ℹ️  No Strigo %s configuration found in %s", strings.ToUpper(sdkType), rcFile)
		return nil
	}

	// Écrire le fichier
	newContent := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(expandedPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update %s: %w", expandedPath, err)
	}

	logging.LogInfo("✅ Successfully removed Strigo %s configuration from %s", strings.ToUpper(sdkType), expandedPath)
	logging.LogInfo("ℹ️  To apply these changes, run: source %s", expandedPath)

	return nil
}

func handleUse(sdkType, distribution, version string) error {
	// Load configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Vérifier si le type de SDK existe
	if _, exists := cfg.SDKTypes[sdkType]; !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Construire le chemin d'installation
	installPath, err := GetInstallPath(cfg, sdkType, distribution, version)
	if err != nil {
		return fmt.Errorf("failed to get installation path: %w", err)
	}

	// Vérifier si la version est installée
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed. Please install it first with:\n  strigo install %s %s %s",
			version, sdkType, distribution, version)
	}

	// Configurer les variables d'environnement
	if sdkType == "jdk" || sdkType == "node" {
		// Trouver le chemin du répertoire bin
		binPath, err := getSDKBinPath(installPath, sdkType)
		if err != nil {
			return err
		}

		// Préparer les exports selon le type de SDK
		var envVar string
		if sdkType == "jdk" {
			envVar = "JAVA_HOME"
		} else {
			envVar = "NODE_HOME"
		}

		exports := fmt.Sprintf("export %s=%s\nexport PATH=$%s/bin:$PATH", envVar, binPath, envVar)

		if setEnvVar {
			rcFile, err := findRcFile()
			if err != nil {
				logging.LogError("❌ Could not find shell configuration file: %v", err)
				logging.LogInfo("ℹ️  Please add these lines manually to your shell configuration:")
				fmt.Println(exports)
				return nil
			}

			// Expand tilde if present
			expandedPath, err := config.ExpandTilde(rcFile)
			if err != nil {
				return fmt.Errorf("failed to expand path %s: %w", rcFile, err)
			}

			// Lire le contenu actuel
			content, err := os.ReadFile(expandedPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", expandedPath, err)
			}

			// Supprimer les anciennes configurations
			lines := strings.Split(string(content), "\n")
			var newLines []string
			for _, line := range lines {
				if !strings.Contains(line, envVar+"=") && !strings.Contains(line, "PATH=$"+envVar) {
					newLines = append(newLines, line)
				}
			}

			// Ajouter les nouvelles configurations avec un commentaire
			newContent := strings.Join(newLines, "\n") + fmt.Sprintf("\n\n# Added by Strigo - %s configuration\n%s\n", strings.ToUpper(sdkType), exports)

			// Écrire le fichier
			if err := os.WriteFile(expandedPath, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to update %s: %w", expandedPath, err)
			}

			logging.LogInfo("✅ Successfully set %s %s version %s as active", sdkType, distribution, version)
			logging.LogInfo("📝 Added to %s:", expandedPath)
			fmt.Println(exports)
			logging.LogInfo("ℹ️  To apply these changes, run: source %s", expandedPath)
		} else {
			logging.LogInfo("✅ Successfully set %s %s version %s as active", sdkType, distribution, version)
			logging.LogInfo("ℹ️  To use this %s, add these lines to your shell configuration:", strings.ToUpper(sdkType))
			fmt.Println(exports)
		}
	}

	return nil
}
