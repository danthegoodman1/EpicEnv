/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get the value of an environment variable",
	Long: `Get the value of an environment variable.

The variable value will be printed to stdout if it exists.
If the variable doesn't exist, the command will exit with status 1.

This is useful for file templating and other scripting scenarios.

Examples:
  epicenv get DATABASE_URL
  epicenv get -e production API_KEY`,
	Run:  runGet,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) {
	key := args[0]
	env := getEnvOrFlag(cmd)
	envMap := loadEnv(env)

	envVar, exists := envMap[key]
	if !exists {
		logger.Error().Msgf("Environment variable '%s' does not exist", key)
		os.Exit(1)
	}

	fmt.Print(envVar.Value)
}
