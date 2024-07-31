package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "epicenv",
	Short: "Epic local environment management in git",
	Long: `Epic local environment management in git.

https://github.com/danthegoodman1/EpicEnv`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("environment", "e", "", "Specify the environment, will use the current by default if one is set")
}
