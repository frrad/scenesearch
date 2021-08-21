package util

import (
	"os/exec"
)

func ExecDebug(x string, args ...string) (string, error) {
	cmd := exec.Command(x, args...)

	z, err := cmd.CombinedOutput()

	return string(z), err
}
