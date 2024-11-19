package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// internalGenCmd represents the internalGen command
var internalGenCmd = &cobra.Command{
	Use:   "zzz_INTERNAL_gen",
	Short: "FOR INTERNAL USE DO NOT RUN: Generates the temporary script to source",
	Run:   runInternalGenCmd,
}

func init() {
	rootCmd.AddCommand(internalGenCmd)
}

func runInternalGenCmd(cmd *cobra.Command, args []string) {
	env := cmd.Flag("environment").Value.String()
	logger.Debug().Msgf("running gen for env %s", env)

	envMap := loadEnv(env)

	// Capture the existing env so we can
	existingEnv := os.Environ()
	existingEnvMap := lo.Associate(existingEnv, func(item string) (string, string) {
		parts := strings.SplitN(item, "=", 2)
		return parts[0], parts[1]
	})

	// generate source file
	var undoLines []string // so we can deactivate the env
	sourceFile := "OLDPS1=$PS1\n"
	for key, loaded := range envMap {
		sourceFile += fmt.Sprintf("export %s=%s\n", key, wrapQuotesIfNeeded(loaded.Value))
		// Check if the env already exists and make a replacement set
		if oldVal, exists := existingEnvMap[key]; exists {
			undoLines = append(undoLines, fmt.Sprintf("export %s=%s\n", key, wrapQuotesIfNeeded(oldVal)))
			continue
		}

		// Otherwise unset the env var
		undoLines = append(undoLines, fmt.Sprintf("unset %s", key))
	}

	// Set the secret env var for introspection
	sourceFile += fmt.Sprintf("export EPICENV=\"%s\"\n", env)
	undoLines = append(undoLines, "unset EPICENV")

	// We allow for nested prefixes so they user knows how they environments are stacked
	sourceFile += fmt.Sprintf("PS1=\"(epicenv: %s)$PS1\"\n", env)
	sourceFile += "epic-deactivate() {\n  "
	sourceFile += strings.Join(undoLines, "\n  ")
	sourceFile += "\n  PS1=$OLDPS1\n"
	sourceFile += "  unset -f epic-deactivate\n"
	sourceFile += "}"

	tempSourcePath := path.Join(".epicenv", env, fmt.Sprintf("temp-%d", time.Now().UnixMilli()))
	err := os.WriteFile(tempSourcePath, []byte(sourceFile), 0777)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing out temp file")
	}

	fmt.Println(tempSourcePath)
}

func wrapQuotesIfNeeded(s string) string {
	// Escape any existing quotes and dollar signs
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "$", "\\$")

	if strings.Contains(s, " ") && s[0:1] != "\"" && s[len(s)-1:] != "\"" {
		return "\"" + s + "\""
	}

	return s
}
