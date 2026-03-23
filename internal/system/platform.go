package system

import (
	"os/exec"
	"runtime"
	"strings"
)

// Platform represents the detected operating system.
type Platform string

const (
	PlatformMacOS   Platform = "macos"
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
)

// DetectPlatform returns the current platform.
func DetectPlatform() Platform {
	switch runtime.GOOS {
	case "darwin":
		return PlatformMacOS
	case "linux":
		return PlatformLinux
	case "windows":
		return PlatformWindows
	default:
		return PlatformLinux
	}
}

// DetectShell returns the current shell name (zsh, bash, fish, powershell).
func DetectShell() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	out, err := exec.Command("basename", "$SHELL").Output()
	if err != nil {
		shell := runCmd("echo", "$SHELL")
		if strings.Contains(shell, "zsh") {
			return "zsh"
		}
		if strings.Contains(shell, "fish") {
			return "fish"
		}
		return "bash"
	}
	return strings.TrimSpace(string(out))
}

// HasCommand checks if a command exists in PATH.
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RunCommand runs a command and returns stdout trimmed.
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runCmd(name string, args ...string) string {
	out, _ := RunCommand(name, args...)
	return out
}
