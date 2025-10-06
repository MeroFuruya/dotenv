# dotenv

A small command-line tool to read `.env` files and output environment variables in a format suitable for various shells.

## Features

- Supports multiple `.env` files and directories
- Recursive search for `.env` files
- Handles comments, quoted values, and multiline values
- Variable interpolation using `${VAR}` syntax
- Outputs in formats compatible with `bash`, `zsh`, `fish`, `powershell`, and `cmd`
- Auto-detects the current shell

## Installation

You can build the tool from source using Go:

```bash
make install
```

## Usage

```bash
dotenv [options]
  -d value
        Directories to search inside (can be specified multiple times) (default: current directory)
  -f value
        Filenames to search for (can be specified multiple times) (default: ".env")
  -q    Suppress non-error output
  -r    Search directories recursively (default: false)
  -s string
        Shell to generate output for (supported: bash, zsh, fish, powershell, cmd, auto-detect, none) (default "auto-detect")
```




