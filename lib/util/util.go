package util

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func ExecDebug(x string, args ...string) (string, error) {
	cmd := exec.Command(x, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return "", nil
}
