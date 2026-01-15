package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetTaskrcPath returns the path to the taskrc file, respecting TASKRC env var and XDG_CONFIG_HOME.
func GetTaskrcPath() string {
	if val := os.Getenv("TASKRC"); val != "" {
		return val
	}
	home, _ := os.UserHomeDir()
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".config")
	}
	return filepath.Join(xdg, "task", "taskrc")
}

// GetTimewConfigPath returns the path to the timewarrior config file, respecting TIMEW_CONFIG env var and XDG_CONFIG_HOME.
func GetTimewConfigPath() string {
	if val := os.Getenv("TIMEW_CONFIG"); val != "" {
		return val
	}
	home, _ := os.UserHomeDir()
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".config")
	}
	return filepath.Join(xdg, "timewarrior", "timewarrior.cfg")
}

// CommandRunner defines the interface for executing commands.
type CommandRunner interface {
	Run(name string, env []string, baseArgs []string, args ...string) (string, error)
}

// DefaultRunner is the standard implementation using os/exec.
type DefaultRunner struct{}

func (r DefaultRunner) Run(name string, env []string, baseArgs []string, args ...string) (string, error) {
	finalArgs := append(baseArgs, args...)
	cmd := exec.Command(name, finalArgs...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	return output, err
}

// Runner is the global command runner used by the application.
var Runner CommandRunner = DefaultRunner{}

// RunCommand executes a command using the global Runner and wraps errors with output.
func RunCommand(name string, env []string, baseArgs []string, args ...string) (string, error) {
	out, err := Runner.Run(name, env, baseArgs, args...)
	if err != nil {
		return "", fmt.Errorf("%s error: %v\nOutput: %s", name, err, out)
	}
	return out, nil
}
