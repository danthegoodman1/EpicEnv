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
		if val.Personal {
			// Add comment for personal vars
			fmt.Printf("%s=%s #personal\n", key, wrapQuotesIfNeeded(val.Value))
		} else {
			fmt.Printf("%s=%s\n", key, wrapQuotesIfNeeded(val.Value))
		}
	}
}
