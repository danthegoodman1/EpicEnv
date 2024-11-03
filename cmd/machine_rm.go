package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var machineRmCmd = &cobra.Command{
	Use:   "machine-rm",
	Short: "Remove a machine user from the environment",
	Run:   runMachineRm,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(machineRmCmd)
}

func runMachineRm(cmd *cobra.Command, args []string) {
	machineName := args[0]
	env := getEnvOrFlag(cmd)

	// Load in the keys
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading keys file")
	}

	// Check if the machine user exists
	if !lo.ContainsBy(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
		return item.MachineName == machineName
	}) {
		logger.Fatal().Msgf("Machine user %s is not invited to this environment", machineName)
	}

	// Load the symmetric key (so we know that we are invited to the env)
	_, err = loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	// Remove the key
	keysFile.EncryptedKeys = lo.Filter(keysFile.EncryptedKeys, func(item EncryptedKey, index int) bool {
		return item.MachineName != machineName
	})

	// Write the updated keys file
	err = writeKeysFile(env, *keysFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing keys file")
	}

	logger.Info().Msgf("Removed machine user %s **THIS IS NOT A REPLACEMENT FOR ROTATING SECRETS!**", machineName)
}
