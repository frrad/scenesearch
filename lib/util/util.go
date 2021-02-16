package util

import (
	"os"
	"os/exec"
)

func ExecDebug(x string, args ...string) (string, error) {
	cmd := exec.Command(x, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return "", nil
}
