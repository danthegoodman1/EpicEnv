package cmd

import (
	"fmt"
	"github.com/samber/lo"
	"os"
	"path"
)

func generateActivateSource(env string) error {
	activateContent := `temp_src=$(%s zzz_INTERNAL_gen -e %s)
if [ $? -lt 1 ]; then
    source $temp_src
    rm $temp_src
fi
`

	err := os.WriteFile(path.Join(".epicenv", env, "activate"), []byte(fmt.Sprintf(activateContent, lo.Ternary(os.Getenv("DEV") == "1", "go run .", "epicenv"), env)), 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}
