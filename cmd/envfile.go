package cmd

import (
	"fmt"
	"strings"

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
		if val.Personal && !val.PersonalSet {
			continue
		}

		if val.Personal {
			// Add comment for personal vars
			fmt.Printf("%s=%s #personal\n", key, formatEnvfileValue(val.Value))
		} else {
			fmt.Printf("%s=%s\n", key, formatEnvfileValue(val.Value))
		}
	}
}

func formatEnvfileValue(s string) string {
	if !strings.Contains(s, `'`) {
		return `'` + s + `'`
	}
	if !strings.Contains(s, `"`) {
		return `"` + s + `"`
	}
	if !strings.Contains(s, "`") {
		return "`" + s + "`"
	}

	escaped := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
	).Replace(s)
	return `"` + escaped + `"`
}
