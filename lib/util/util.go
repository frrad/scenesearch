package util

import (
	"io/ioutil"
	"os/exec"
)

func ExecDebug(x string, args ...string) (string, error) {
	cmd := exec.Command(x, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	slurp, _ := ioutil.ReadAll(stderr)

	errStr := string(slurp)

	if err := cmd.Wait(); err != nil {
		return errStr, err
	}

	return errStr, nil
}
