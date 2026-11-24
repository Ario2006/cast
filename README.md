# cast

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)]()
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)]()

**`cast`** is a fast, local-first command-line utility for developers and power users. It simplifies sharing files and live dev servers across devices on your local network, keeps a secure local clipboard history, evaluates arithmetic expressions without shell overhead, and performs instant unit conversions with an interactive terminal picker.

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Homebrew (macOS & Linux)](#homebrew-macos--linux)
  - [Go Install](#go-install)
  - [Building from Source](#building-from-source)
  - [Shell Completion](#shell-completion)
- [Quick Reference](#quick-reference)
- [Use Cases & Command Guide](#use-cases--command-guide)
  - [1. Local Network Sharing](#1-local-network-sharing)
    - [Live Project & Dev Server Hosting (`cast live`)](#live-project--dev-server-hosting-cast-live)
    - [Connecting to a Live Session (`cast connect`)](#connecting-to-a-live-session-cast-connect)
    - [File & Folder Sharing (`cast share`)](#file--folder-sharing-cast-share)
    - [Receiving Shared Files (`cast receive`)](#receiving-shared-files-cast-receive)
  - [2. Productivity Utilities](#2-productivity-utilities)
    - [Clipboard History Picker (`cast clip`)](#clipboard-history-picker-cast-clip)
    - [Safe Shell Calculator (`cast calc`)](#safe-shell-calculator-cast-calc)
    - [Interactive Unit Conversion (`cast conv`)](#interactive-unit-conversion-cast-conv)
- [Supported Units](#supported-units)
- [Configuration](#configuration)
- [Security & Privacy](#security--privacy)
- [Uninstallation](#uninstallation)
- [Development](#development)

---

## Features

- 🌐 **Zero-Config LAN Sharing**: Expose local folders or proxy running dev servers (e.g. Next.js, Vite, Rails) to mobile devices or other computers via short codes or terminal QR codes.
- 📦 **On-the-Fly Directory Streaming**: Directories shared via `cast share` are automatically packaged as streaming ZIP archives and cleaned up on expiration.
- 📋 **Local Clipboard History**: SQLite-backed clipboard picker with sensitive token redaction and keyboard navigation.
- 🧮 **Safe Arithmetic Parser**: Decimal expression evaluation with zero shell execution or `eval` vulnerabilities.
- ⚖️ **Interactive Unit Conversion**: Convert length, mass, temperature, time, data size, and speed with an interactive Bubble Tea terminal menu or direct CLI arguments.
- 🔒 **Private & Local-First**: No accounts, no cloud relays, no tracking telemetry. All network traffic stays strictly on your local network with Ed25519 device identities.

---

## Installation

### Homebrew (macOS & Linux)

```bash
brew install Ario2006/tap/cast
```

### Go Install

```bash
go install github.com/aryankumar/cast/cmd/cast@latest
```

### Building from Source

Ensure Go 1.22+ is installed:

```bash
git clone https://github.com/aryankumar/cast.git
cd cast
make build
# Binary is built at ./cast
```

### Shell Completion

Generate command autocompletion for your shell:

```bash
# Bash
cast completion bash > /etc/bash_completion.d/cast

# Zsh
cast completion zsh > "${fpath[1]}/_cast"

# Fish
cast completion fish > ~/.config/fish/completions/cast.fish

# PowerShell
cast completion powershell | Out-String | Invoke-Expression
```

---

## Quick Reference

| Command | Arguments | Description |
| :--- | :--- | :--- |
| `cast live` | `<path>` or `<port>` | Expose a static directory or reverse-proxy a running local port over LAN |
| `cast connect` | `<code>` | Resolve and connect to an active live session by short code |
| `cast share` | `<file>` or `<dir>` | Share a file or folder on the local network (10-minute expiry) |
| `cast receive` | `<code>` | Download a shared file or directory archive by short code |
| `cast clip` | _none_ | Launch interactive TUI clipboard history picker |
| `cast clip clear` | _none_ | Delete all stored local clipboard history |
| `cast calc` | `<expr>` | Evaluate an arithmetic expression |
| `cast conv` | `<val> <unit> [target]`| Convert between units (interactive picker if target omitted) |
| `cast completion` | `<shell>` | Generate auto-completion script (`bash`, `zsh`, `fish`, `powershell`) |
| `cast version` | _none_ | Display current version |

**Global Flags:**
- `--debug`: Enable verbose diagnostic logging on `stderr`.
- `-h`, `--help`: Display help information for any command.
- `-v`, `--version`: Print `cast` version.

---

## Use Cases & Command Guide

### 1. Local Network Sharing

#### Live Project & Dev Server Hosting (`cast live`)

Quickly test responsive designs on your phone or share progress with a colleague on the same Wi-Fi.

```bash
# Host the current working directory as a read-only HTTP file server
cast live .

# Host a specific project folder
cast live ./my-static-site

# Reverse-proxy a local development server running on port 3000
cast live 3000
```

**How it works:**
- Spawns a lightweight local HTTP server and broadcasts its presence on the local network via UDP (`39422`).
- Prints an ASCII QR code for phone cameras and outputs a human-readable 6-character code (e.g. `7F92K8`).
- Automatically shuts down after 10 minutes or when pressing `Ctrl+C`.

#### Connecting to a Live Session (`cast connect`)

From another computer on the same network:

```bash
cast connect 7F92K8
```

Resolves the host's LAN IP and session information, displaying the device name and remote URL.

#### File & Folder Sharing (`cast share`)

Send large files or entire directory trees directly across machines without uploading to third-party cloud drives.

```bash
# Share a single document or media file
cast share presentation.pdf

# Share a folder (streamed as a ZIP archive on the fly)
cast share ./assets
```

- Renders a terminal QR code for direct browser downloads.
- Outputs a 6-character code for receiver CLI downloads.
- Automatically cleans up temporary ZIP archives and expires after 10 minutes.

#### Receiving Shared Files (`cast receive`)

Download files shared with `cast share`:

```bash
cast receive 7F92K8
```

- Displays real-time download percentage and transfer progress.
- **Safety guarantee:** Will never silently overwrite an existing file in your current directory.

---

### 2. Productivity Utilities

#### Clipboard History Picker (`cast clip`)

A private, lightweight clipboard manager right in your terminal.

```bash
# Open the interactive history selector
cast clip

# Wipe local clipboard history database
cast clip clear
```

**Features:**
- **Interactive Navigation:** Use `↑` / `↓` or type number indices to navigate, `Enter` to restore an entry to your system clipboard, and `q` or `Esc` to exit.
- **Privacy Filter:** Automatically analyzes clipboard text and avoids saving sensitive secrets (passwords, tokens, high-entropy keys).
- **Local SQLite Store:** Stored locally in SQLite with a default 200-item retention limit. Never synced off-device.

#### Safe Shell Calculator (`cast calc`)

Perform quick math without launching Python, Node, or `bc`.

```bash
cast calc 15 * 32
cast calc '(12 + 8) * 3'
cast calc 100 / 4 + 2.5
cast calc 2 ^ 10
cast calc 27 % 4
```

- **Supported operators:** Addition (`+`), subtraction (`-`), multiplication (`*`), division (`/`), modulo (`%`), exponentiation (`^`), unary minus/plus, and nested parentheses (`(...)`).
- **Safe:** Evaluated with a dedicated recursive-descent parser. Zero shell evaluation or code injection risks.

#### Interactive Unit Conversion (`cast conv`)

Convert measurements instantly using either an interactive terminal menu or direct CLI arguments.

```bash
# Interactive mode (opens a Bubble Tea selection menu of compatible units)
cast conv 10 km
cast conv 72 mph
cast conv 98.6 F

# Direct conversion mode
cast conv 10 km mi
cast conv 100 C F
cast conv 2.5 GB MB
cast conv 80 kg lb
```

---

## Supported Units

All conversions run 100% offline. Data units use standard decimal (base-10) notation:

| Category | Supported Symbols | Description |
| :--- | :--- | :--- |
| **Length** | `mm`, `cm`, `m`, `km`, `in`, `ft`, `yd`, `mi` | Millimeters, centimeters, meters, kilometers, inches, feet, yards, miles |
| **Mass** | `mg`, `g`, `kg`, `oz`, `lb` | Milligrams, grams, kilograms, ounces, pounds |
| **Temperature** | `C`, `F`, `K` | Celsius, Fahrenheit, Kelvin |
| **Time** | `ms`, `s`, `min`, `h`, `day` | Milliseconds, seconds, minutes, hours, days |
| **Data Size** | `B`, `KB`, `MB`, `GB`, `TB` | Bytes, kilobytes (10³), megabytes (10⁶), gigabytes (10⁹), terabytes (10¹²) |
| **Speed** | `m/s`, `km/h`, `mph`, `ft/s` | Meters/sec, km/h, miles per hour, feet/sec |

---

## Configuration

`cast` requires no configuration to run. Optional configuration can be placed in your platform configuration directory:

- **macOS:** `~/Library/Application Support/cast/config.toml` (or `~/.config/cast/config.toml`)
- **Linux:** `~/.config/cast/config.toml`
- **Windows:** `%APPDATA%\cast\config.toml`

### Example `config.toml`

```toml
[clipboard]
max_history = 200      # Number of clipboard entries to retain

[share]
expires_minutes = 10   # Session lifetime for share and live commands

[network]
preferred_port = 0     # 0 chooses an available OS port automatically
```

### Environment Variables

| Variable | Description |
| :--- | :--- |
| `CAST_CONFIG_DIR` | Override the default directory for configuration, SQLite database, and device keys (useful for testing or scripting). |

---

## Security & Privacy

- **Local Network Boundary:** No remote cloud relay or third-party servers. All transfers occur point-to-point over the local network.
- **Cryptographic Device Identity:** An Ed25519 keypair and device fingerprint are generated on first launch and stored with restricted permissions (`0600`).
- **Tokenized Authorization:** URLs and QR codes contain high-entropy authorization secrets preventing brute-force access from other local devices.
- **Path Traversal Protection:** Directory hosting is strictly confined to the specified root folder; symlinks attempting to escape are blocked.
- **Sensitive Content Shield:** Clipboard monitoring ignores high-entropy secrets and sensitive credential patterns to avoid saving passwords or private keys.

---

## Uninstallation

To completely remove `cast` and its local data from your machine:

### 1. Remove the Binary

**Homebrew:**
```bash
brew uninstall cast
brew untap Ario2006/tap  # optional
```

**Go Install:**
```bash
rm -f $(go env GOPATH)/bin/cast
```

**Manual Binary (in PATH):**
```bash
rm -f $(which cast)
```

### 2. Remove Configuration & Local Data

To delete the SQLite clipboard database (`history.db`), cryptographic identity keys, and configuration:

```bash
# macOS
rm -rf ~/Library/Application\ Support/cast ~/.config/cast

# Linux
rm -rf ~/.config/cast

# Windows (PowerShell)
Remove-Item -Recurse -Force "$env:APPDATA\cast"
```

---

## Development

`cast` is written in Go using standard build tools.

```bash
# Format source code
make fmt

# Run static analysis
make vet

# Run all unit and integration tests
make test

# Compile local binary
make build
```

---

## License

This project is licensed under the MIT License.
