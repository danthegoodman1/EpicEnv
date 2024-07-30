package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// internalGenCmd represents the internalGen command
var internalGenCmd = &cobra.Command{
	Use:   "zzz_INTERNAL_gen",
	Short: "FOR INTERNAL USE DO NOT RUN: Generates the temporary script to source",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("zzz_INTERNAL_gen called")
	},
}

func init() {
	rootCmd.AddCommand(internalGenCmd)
}
