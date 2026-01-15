package common

import (
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner defines the interface for executing commands.
type CommandRunner interface {
	Run(name string, baseArgs []string, args ...string) (string, error)
}

// DefaultRunner is the standard implementation using os/exec.
type DefaultRunner struct{}

func (r DefaultRunner) Run(name string, baseArgs []string, args ...string) (string, error) {
	finalArgs := append(baseArgs, args...)
	cmd := exec.Command(name, finalArgs...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		return "", fmt.Errorf("%s error: %v\nOutput: %s", name, err, output)
	}
	return output, nil
}

// Runner is the global command runner used by the application.
var Runner CommandRunner = DefaultRunner{}

// RunCommand executes a command using the global Runner.
func RunCommand(name string, baseArgs []string, args ...string) (string, error) {
	return Runner.Run(name, baseArgs, args...)
}
