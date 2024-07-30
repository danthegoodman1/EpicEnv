/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// inviteCmd represents the invite command
var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a GitHub user to the EpicEnv, allowing them to decrypt the environment",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("invite called")
	},
}

func init() {
	rootCmd.AddCommand(inviteCmd)
}
