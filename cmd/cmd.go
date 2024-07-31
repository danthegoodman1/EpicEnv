package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/danthegoodman1/epicenv/gologger"
	"golang.org/x/term"
	"os"
	"path"
	"strings"
	"syscall"
)

var (
	logger = gologger.NewLogger()
)

func readStdinHidden(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(syscall.Stdin)
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
