package cmd

import (
	"github.com/spf13/cobra"
)

// integrateCmd represents the integrate command
var integrateCmd = &cobra.Command{
	Use:   "integrate INTEGRATION",
	Short: "Add an integration to EpicEnv",
	Long: `Add an integration to EpicEnv

Available integrations: 1password`,
	Args: cobra.ExactArgs(1),
	Run:  runIntegrate,
}

func init() {
	rootCmd.AddCommand(integrateCmd)
}

func runIntegrate(cmd *cobra.Command, args []string) {
	integration := Integration(args[0])
	env := getEnvOrFlag(cmd)

	// Verify we are in the right directory
	if !envExists(env) {
		logger.Fatal().Msgf("Environment %s not found, make sure there is a .epicenv directory in this directory", env)
	}

	switch integration {
	case Integration1Password:
		err := Setup1PasswordServiceAccount(env, readStdinHidden("Insert your 1Password Service Account Token> "))
		if err != nil {
			logger.Fatal().Err(err).Msg("Error setting up 1Password service account")
		}
	default:
		logger.Fatal().Msgf("Unknown integration \"%s\"", integration)
	}
}
