package cmd

import (
	"github.com/spf13/cobra"
)

// internalGenCmd represents the internalGen command
var internalGenCmd = &cobra.Command{
	Use:   "zzz_INTERNAL_gen",
	Short: "FOR INTERNAL USE DO NOT RUN: Generates the temporary script to source",
	Run:   runInternalGenCmd,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(internalGenCmd)
}

func runInternalGenCmd(cmd *cobra.Command, args []string) {
	env := args[0]
	logger.Debug().Msgf("running gen for env %s", env)
}
