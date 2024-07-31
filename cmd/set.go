/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set an environment variable",
	Long: `Set an environment variable

Use -e to set the environment, can omit to use the current.

Use -p to set a personal variable.

If you attempt to normal set a personal variable, it will update the personal variable instead. To make a personal variable shared, first rm the variable, then set it again as shared.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("set called")
	},
	Args:       cobra.RangeArgs(2, 3),
	ArgAliases: []string{"env", "key", "value"},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().BoolP("personal", "p", false, "Set this as a personal environment if it doesn't exist")
	setCmd.Flags().BoolP("stdin", "i", false, "Read the value from stdin instead")
}
