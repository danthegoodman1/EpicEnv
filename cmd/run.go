package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [--] [command] [args...]",
	Short: "Run a command with environment variables injected",
	Long: `Run a command with the specified environment's variables injected.

The command will inherit all current environment variables, with epicenv
variables added/overriding them.

Use -- to separate epicenv flags from the command's flags.

Examples:
  epicenv run go test ./...
  epicenv -e staging run ./my-binary
  epicenv -e production run bash -c 'echo $DATABASE_URL'
  epicenv run -- ./my-binary --help`,
	Run:  runRun,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) {
	env := getEnvOrFlag(cmd)
	envMap := loadEnv(env)

	// Build environment: start with current env, then add epicenv vars
	environ := os.Environ()
	for key, val := range envMap {
		if val.Value != "" {
			environ = append(environ, key+"="+val.Value)
		}
	}

	// Create the command
	execCmd := exec.Command(args[0], args[1:]...)
	execCmd.Env = environ
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Forward signals to the child process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := execCmd.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start command")
	}

	go func() {
		for sig := range sigChan {
			if execCmd.Process != nil {
				execCmd.Process.Signal(sig)
			}
		}
	}()

	if err := execCmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		logger.Fatal().Err(err).Msg("command failed")
	}
}
