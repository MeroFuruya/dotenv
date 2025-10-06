# dotenv

[![Build and Release](https://github.com/MeroFuruya/dotenv/actions/workflows/build.yml/badge.svg)](https://github.com/MeroFuruya/dotenv/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/MeroFuruya/dotenv)](https://goreportcard.com/report/github.com/MeroFuruya/dotenv)
[![License](https://img.shields.io/github/license/MeroFuruya/dotenv)](LICENSE)

A small command-line tool to read `.env` files and output environment variables in a format suitable for various shells.

## Features

- Supports multiple `.env` files and directories
- Recursive search for `.env` files
- Handles comments, quoted values, and multiline values
- Variable interpolation using `${VAR}` syntax
- Outputs in formats compatible with `bash`, `zsh`, `fish`, `powershell`, and `cmd`
- Auto-detects the current shell

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/MeroFuruya/dotenv/releases).

Available platforms:
- **Linux**: amd64, arm64, arm-v7, 386
- **Windows**: amd64, arm64, 386
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **FreeBSD**: amd64, arm64
- **OpenBSD**: amd64, arm64
- **NetBSD**: amd64, arm64

### Docker

```bash
# Pull the latest image
docker pull ghcr.io/merofuruya/dotenv:latest

# Run with mounted .env file
docker run --rm -v $(pwd)/.env:/app/.env ghcr.io/merofuruya/dotenv:latest -d /app
```

### Build from Source

Requirements: Go 1.22.4 or later

```bash
# Clone the repository
git clone https://github.com/MeroFuruya/dotenv.git
cd dotenv

# Build for your current platform
go build -o dotenv .

# Or use the Makefile
make install
```

### Cross-compilation

Build for different platforms:

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o dotenv-linux-amd64 .

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o dotenv-windows-amd64.exe .

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dotenv-darwin-arm64 .
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




