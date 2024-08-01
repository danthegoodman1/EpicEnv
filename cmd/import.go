package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import PATH",
	Short: "Import and existing env var",
	Args:  cobra.ExactArgs(1),
	Run:   runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().BoolP("override", "o", false, "Override existing values")
}

func runImport(cmd *cobra.Command, args []string) {
	envPath := args[0]
	env := getEnvOrFlag(cmd)

	// Read in the file
	fileContent, err := os.ReadFile(envPath)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error reading %s", envPath)
	}

	lines := strings.Split(string(fileContent), "\n")
	lines = lo.Filter(lines, func(item string, index int) bool {
		return item != "" && len(strings.SplitN(item, "=", 2)) == 2
	})

	loadedEnvMap := lo.Associate(lines, func(item string) (string, string) {
		parts := strings.SplitN(item, "=", 2)
		return parts[0], parts[1]
	})

	logger.Debug().Interface("loadedEnvVars", lo.Keys(loadedEnvMap)).Msg("loaded env map")

	for key, val := range loadedEnvMap {
		setEnvVar(cmd, env, key, val)
	}

	logger.Info().Msgf("Imported %d variables from %s", len(loadedEnvMap), envPath)
}
