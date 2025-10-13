package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type ArrayFlags []string

// String is an implementation of the flag.Value interface
func (i *ArrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

// Set is an implementation of the flag.Value interface
func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var quiet bool

func main() {
	var dir ArrayFlags
	flag.Var(&dir, "d", "Directories to search inside (can be specified multiple times) (default: current directory)")
	var name ArrayFlags
	flag.Var(&name, "f", "Filenames to search for (can be specified multiple times) (default: \".env\")")
	var recursive bool
	flag.BoolVar(&recursive, "r", false, "Search directories recursively (default: false)")
	var shell string
	flag.StringVar(&shell, "s", "auto-detect", "Shell to generate output for (supported: bash, zsh, fish, powershell, cmd, auto-detect, none)")
	flag.BoolVar(&quiet, "q", false, "Suppress non-error output")
	var filter string
	flag.StringVar(&filter, "filter", "", "Only output variables that match this regex pattern")
	flag.Parse()

	if len(dir) == 0 {
		dir = append(dir, ".")
	}
	if len(name) == 0 {
		name = append(name, ".env")
	}

	dotenvFile := SearchFile(dir, name, recursive)
	if dotenvFile != "" {
		Log("Using dotenv file:", dotenvFile)
	} else {
		Error("No dotenv file found")
	}

	parser := NewParser()
	envMap, err := parser.ParseFile(dotenvFile)
	if err != nil {
		Error("Error reading dotenv file:", err)
		return
	}

	if shell == "auto-detect" {
		shell = DetectShell()
		Log("Auto-detected shell:", shell)
	}

	lines := []string{}

	for _, variable := range envMap {
		if filter != "" {
			matched, err := MatchRegex(filter, variable.Name)
			if err != nil {
				Error("Invalid regex pattern:", filter, err)
				return
			}
			if !matched {
				continue
			}
		}

		line := TransformToShellSyntax(variable, shell)
		if line != "" {
			lines = append(lines, line)
		}
	}
	fmt.Print(strings.Join(lines, "\n"))
}

func SearchFile(directories, names []string, recursive bool) string {
	for _, dir := range directories {
		files, err := os.ReadDir(dir)
		if err != nil && os.IsExist(err) {
			continue
		} else if err != nil {
			Error("Error reading directory:", dir, err)
			continue
		}

		for _, name := range names {
			for _, file := range files {
				if file.Name() == name && !file.IsDir() {
					return fmt.Sprintf("%s/%s", dir, name)
				}
			}
		}

		if recursive {
			found := SearchFileInSubdirs(dir, names)
			if found != "" {
				return found
			}
		}
	}
	return ""
}

func SearchFileInSubdirs(directory string, names []string) string {
	entries, err := os.ReadDir(directory)
	if err != nil {
		Error("Error reading directory:", directory, err)
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subdir := fmt.Sprintf("%s/%s", directory, entry.Name())
			for _, name := range names {
				subEntries, err := os.ReadDir(subdir)
				if err != nil {
					continue
				}
				for _, subEntry := range subEntries {
					if subEntry.Name() == name && !subEntry.IsDir() {
						return fmt.Sprintf("%s/%s", subdir, name)
					}
				}
			}
			found := SearchFileInSubdirs(subdir, names)
			if found != "" {
				return found
			}
		}
	}
	return ""
}

func Log(message ...any) {
	if !quiet {
		fmt.Fprintln(os.Stderr, append([]any{"[Log]"}, message...)...)
	}
}

func Error(message ...any) {
	fmt.Fprintln(os.Stderr, append([]any{"[Error]"}, message...)...)
}

func MatchRegex(pattern, str string) (bool, error) {
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, err
	}
	return matched, nil
}
