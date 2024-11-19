package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/samber/lo"
)

func generateActivateSource(env string) error {
	activateContent := `if [ -z "${EPICENV}" ]; then
    # do nothing
else
    echo deactivating $EPICENV
    epic-deactivate
fi
temp_src=$(%s zzz_INTERNAL_gen -e %s)
if [ $? -lt 1 ]; then
	echo activated env %s
    source $temp_src
    if [ -z "${EPICENV_DEV}" ]; then
        rm $temp_src
    else
        echo leaving temp file $temp_src for debug
    fi
fi
`

	err := os.WriteFile(path.Join(".epicenv", env, "activate"), []byte(fmt.Sprintf(activateContent, lo.Ternary(os.Getenv("EPICENV_DEV") != "", "go run .", "epicenv"), env, env)), 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}
