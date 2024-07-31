package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize EpicEnv",

	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("init called")
	// todo name env (default)
	// todo check if name already exists
	// todo collect their keys from github, verify they exist in $HOME/.ssh/
	// todo create initial symmetric key and encrypt it with their key
	// todo write keys to disk
	// todo create source script
	// TODO: Record to audit log
}
