package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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

func main() {
	var dir ArrayFlags
	flag.Var(&dir, "d", "Directories to search for dotenv files (can be specified multiple times) (default: current directory)")
	var name ArrayFlags
	flag.Var(&name, "n", "Names of dotenv files to search for (can be specified multiple times) (default: \".env\")")
	flag.Parse()

	if len(dir) == 0 {
		dir = append(dir, ".")
	}
	if len(name) == 0 {
		name = append(name, ".env")
	}

	dotenvFile := SearchFile(dir, name)
	if dotenvFile != "" {
		fmt.Fprintln(os.Stderr, "Using dotenv file:", dotenvFile)
	} else {
		fmt.Fprintln(os.Stderr, "No dotenv file found")
	}

	parser := NewParser()
	envMap, err := parser.ParseFile(dotenvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading dotenv file: %v\n", err)
		return
	}

	for _, variable := range envMap {
		fmt.Println(VariableToBash(variable))
	}
}

func SearchFile(directories, names []string) string {
	for _, dir := range directories {
		files, err := os.ReadDir(dir)
		if err != nil && os.IsExist(err) {
			continue
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dir, err)
			continue
		}

		for _, name := range names {
			for _, file := range files {
				if file.Name() == name && !file.IsDir() {
					return fmt.Sprintf("%s/%s", dir, name)
				}
			}
		}
	}
	return ""
}

func VariableToBash(variable Variable) string {
	escapedValue := strconv.Quote(variable.Value)
	return fmt.Sprintf("export %s=%s", variable.Name, escapedValue)
}
