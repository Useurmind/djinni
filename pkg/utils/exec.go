package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func ExecCommand(name string, args []string, workdir string) error {
	_, err := ExecCommandWithOutput(name, args, workdir)
	return err
}

func ExecCommandWithOutput(name string, args []string, workdir string) (string, error) {
	cmd := exec.Command(name, args...)
	if workdir != "" {
		cmd.Dir = workdir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s failed: %v, output: %s", name, err, string(output))
	}

	if name == "git" && strings.Contains(string(output), "fatal:") {
		return "", fmt.Errorf("git command failed: %s", string(output))
	}

	return string(output), nil
}
