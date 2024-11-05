package cmd

import (
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var machineAddCmd = &cobra.Command{
	Use:   "machine-add",
	Short: "Add a machine user to the EpicEnv using their public key",
	Long: `Add a machine user (like production/staging servers) to the EpicEnv using their public key.
Example: epicenv machine-add prod-server /path/to/public_key.pub`,
	Run:  runMachineAdd,
	Args: cobra.ExactArgs(2),
}

func init() {
	rootCmd.AddCommand(machineAddCmd)
}

func runMachineAdd(cmd *cobra.Command, args []string) {
	machineName := args[0]
	pubKeyPath := args[1]
	env := getEnvOrFlag(cmd)

	// Read the public key file
	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading public key file")
	}
	pubKey := string(pubKeyBytes)

	// Load existing keys
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading keys file")
	}

	// Check if machine name already exists
	if lo.ContainsBy(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
		return item.MachineName == machineName
	}) {
		logger.Fatal().Msgf("Machine user %s already exists", machineName)
	}

	// Load symmetric key
	symKey, err := loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	// Encrypt symmetric key with their public key
	encSymKey, err := encryptWithPublicKey(symKey, pubKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("error encrypting with public key")
	}

	// Add the new machine user
	encKey := EncryptedKey{
		PublicKey:          pubKey,
		EncryptedSharedKey: encSymKey,
		MachineName:        machineName,
	}
	keysFile.EncryptedKeys = append(keysFile.EncryptedKeys, encKey)

	// Write the updated keys file
	err = writeKeysFile(env, *keysFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing keys file")
	}

	logger.Info().Msgf("Added machine user %s", machineName)
}
