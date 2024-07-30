/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// uninviteCmd represents the uninvite command
var uninviteCmd = &cobra.Command{
	Use:   "uninvite",
	Short: "Uninvite a GitHub collaborator from the environment, removing their access by re-encrypting. This is not a replacement for rotating secrets!",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("uninvite called")
	},
}

func init() {
	rootCmd.AddCommand(uninviteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uninviteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uninviteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
