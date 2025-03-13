package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// envfileCmd represents the envfile command
var envfileCmd = &cobra.Command{
	Use:   "envfile",
	Short: "Export environment as .env file contents to stdout",
	Long: `Export environment as .env file contents to stdout.
	
Example:
  epicenv envfile -e prod > .env`,
	Run: runEnvfile,
}

func init() {
	rootCmd.AddCommand(envfileCmd)
}

func runEnvfile(cmd *cobra.Command, args []string) {
	env := getEnvOrFlag(cmd)
	envMap := loadEnv(env)

	// Output in .env format
	for key, val := range envMap {
		// Escape backslashes in the value
		escapedValue := escapeBackslashes(val.Value)

		if val.Personal {
			// Add comment for personal vars
			fmt.Printf("%s=%s #personal\n", key, wrapQuotesIfNeeded(escapedValue))
		} else {
			fmt.Printf("%s=%s\n", key, wrapQuotesIfNeeded(escapedValue))
		}
	}
}

// escapeBackslashes replaces each backslash with three backslashes because bash
func escapeBackslashes(s string) string {
	// Replace each \ with \\
	var result string
	for _, c := range s {
		if c == '\\' {
			result += "\\\\\\"
		} else {
			result += string(c)
		}
	}
	return result
}
