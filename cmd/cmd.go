package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/danthegoodman1/epicenv/gologger"
)

var (
	logger = gologger.NewLogger()
)

func readStdinHidden(prompt string) string {
	fmt.Print(prompt)
	// IDE might complain, but the cast is necessary for some OSs, because Stdin is a var instead of an untyped const
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println() // Move to the next line after input
	return string(bytePassword)
}

func readStdin(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// Optimistic check, does not check for permissions
func envExists(env string) bool {
	if _, err := os.Stat(path.Join(".epicenv", env)); errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		logger.Fatal().Err(err).Msg("error checking if file exists")
	}

	return true
}
