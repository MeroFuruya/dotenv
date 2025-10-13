package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func IsShell(name string) bool {
	shell := os.Getenv("SHELL")
	return strings.HasSuffix(shell, name)
}

func DetectShell() string {
	shells := []string{"bash", "zsh", "fish", "powershell", "sh"}
	for _, shell := range shells {
		if IsShell(shell) {
			return shell
		}
	}
	if os.Getenv("COMSPEC") != "" || IsShell("pwsh") {
		return "powershell"
	}
	if IsShell("cmd.exe") {
		return "cmd"
	}
	return ""
}

func TransformToShellSyntax(variable Variable, shellName string) string {
	value := strconv.Quote(variable.Value)
	switch shellName {
	case "bash", "zsh", "sh":
		return fmt.Sprintf("export %s=%s", variable.Name, value)
	case "fish":
		return fmt.Sprintf("set -x %s %s", variable.Name, value)
	case "cmd":
		return fmt.Sprintf("set %s=%s", variable.Name, value)
	case "powershell":
		return fmt.Sprintf("$env:%s=%s", variable.Name, value)
	case "none":
		return fmt.Sprintf("%s=%q", variable.Name, variable.Value)
	case "value":
		return variable.Value
	default:
		return ""
	}
}
