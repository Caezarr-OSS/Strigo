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
)

func init() {
	useCmd.Flags().BoolVarP(&setEnvVar, "set-env", "e", false, "Set environment variables in shell configuration file (~/.bashrc or ~/.zshrc)")
}

var useCmd = &cobra.Command{
	Use:   "use [type] [distribution] [version]",
	Short: "Set a specific SDK version as active",
	Long: `Set a specific SDK version as active. For example:
strigo use jdk temurin 11.0.24_8

This will create a symbolic link to the specified version.`,
	Args: func(cmd *cobra.Command, args []string) error {
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
	if err := handleUse(args[0], args[1], args[2]); err != nil {
		ExitWithError(err)
	}
}

func getJDKBinPath(basePath string) (string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	var jdkDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "jdk") {
			jdkDir = entry.Name()
			break
		}
	}

	if jdkDir == "" {
		return "", fmt.Errorf("could not find JDK directory in %s", basePath)
	}

	return filepath.Join(basePath, jdkDir), nil
}

func findRcFile() (string, error) {
	shell := os.Getenv("SHELL")
	home := os.Getenv("HOME")
	
	// Liste des fichiers RC possibles dans l'ordre de préférence
	rcFiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
	}

	// Chercher le premier fichier RC existant
	for _, file := range rcFiles {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("no shell configuration file found (.zshrc or .bashrc)")
}

func handleUse(sdkType, distribution, version string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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

	// Configurer JAVA_HOME pour les JDKs
	if sdkType == "jdk" {
		jdkPath, err := getJDKBinPath(installPath)
		if err != nil {
			return err
		}

		// Préparer les exports
		exports := fmt.Sprintf("export JAVA_HOME=%s\nexport PATH=$JAVA_HOME/bin:$PATH", jdkPath)

		if setEnvVar {
			rcFile, err := findRcFile()
			if err != nil {
				logging.LogError("❌ Could not find shell configuration file: %v", err)
				logging.LogInfo("ℹ️  Please add these lines manually to your shell configuration:")
				fmt.Println(exports)
				return nil
			}

			// Lire le contenu actuel
			content, err := os.ReadFile(rcFile)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", rcFile, err)
			}

			// Supprimer les anciennes configurations JAVA_HOME
			lines := strings.Split(string(content), "\n")
			var newLines []string
			for _, line := range lines {
				if !strings.Contains(line, "JAVA_HOME=") && !strings.Contains(line, "PATH=$JAVA_HOME") {
					newLines = append(newLines, line)
				}
			}

			// Ajouter les nouvelles configurations avec un commentaire
			newContent := strings.Join(newLines, "\n") + "\n\n# Added by Strigo - JDK configuration\n" + exports + "\n"

			// Écrire le fichier
			if err := os.WriteFile(rcFile, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to update %s: %w", rcFile, err)
			}

			logging.LogInfo("✅ Successfully set %s %s version %s as active", sdkType, distribution, version)
			logging.LogInfo("📝 Added to %s:", rcFile)
			fmt.Println(exports)
			logging.LogInfo("ℹ️  To apply these changes, run: source %s", rcFile)
		} else {
			logging.LogInfo("✅ Successfully set %s %s version %s as active", sdkType, distribution, version)
			logging.LogInfo("ℹ️  To use this JDK, add these lines to your shell configuration:")
			fmt.Println(exports)
		}
	}

	return nil
}
