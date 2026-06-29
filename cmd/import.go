package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
	"unicode"
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

	loadedEnvMap := map[string]importedEnvfileVar{}
	for _, line := range strings.Split(string(fileContent), "\n") {
		key, val, personal, ok := parseEnvfileLine(line)
		if !ok {
			continue
		}
		loadedEnvMap[key] = importedEnvfileVar{
			value:    val,
			personal: personal,
		}
	}

	for key, val := range loadedEnvMap {
		if val.personal {
			logger.Info().Msgf("importing \"%s\" as personal", key)
		}
		setEnvVar(env, key, val.value, val.personal)
	}

	logger.Info().Msgf("Imported %d variables from %s", len(loadedEnvMap), envPath)
}

type importedEnvfileVar struct {
	value    string
	personal bool
}

func parseEnvfileLine(line string) (string, string, bool, bool) {
	if line == "" {
		return "", "", false, false
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false, false
	}

	val, personal := parseEnvfileValue(parts[1])
	return parts[0], val, personal, true
}

func parseEnvfileValue(s string) (string, bool) {
	if strings.HasPrefix(s, `"`) || strings.HasPrefix(s, `'`) || strings.HasPrefix(s, "`") {
		return parseQuotedEnvfileValue(s)
	}

	trimmed := strings.TrimRightFunc(s, unicode.IsSpace)
	if strings.HasSuffix(trimmed, "#personal") {
		markerStart := len(trimmed) - len("#personal")
		if markerStart > 0 && unicode.IsSpace(rune(trimmed[markerStart-1])) {
			return strings.TrimSpace(trimmed[:markerStart]), true
		}
	}

	return s, false
}

func parseQuotedEnvfileValue(s string) (string, bool) {
	quote := s[0]
	end := strings.LastIndex(s, string(quote))
	if end > 0 {
		value := s[1:end]
		if quote == '"' && strings.Contains(value, `\"`) {
			value = strings.NewReplacer(
				`\\`, `\`,
				`\"`, `"`,
			).Replace(value)
		}
		return value, strings.TrimSpace(s[end+1:]) == "#personal"
	}
	return s, false
}
