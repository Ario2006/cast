# cast

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)]()
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)]()

**`cast`** is a fast, local-first command-line utility for developers and power users. It simplifies sharing files and live dev servers across devices on your local network, keeps a private local clipboard history, evaluates arithmetic expressions without shell overhead, and performs instant unit conversions with an interactive terminal picker.

---

## Table of Contents

- [Core Principles](#core-principles)
- [Installation](#installation)
  - [Homebrew (macOS & Linux)](#homebrew-macos--linux)
  - [Go Install](#go-install)
  - [Building from Source](#building-from-source)
  - [Shell Completion](#shell-completion)
- [Quick Reference](#quick-reference)
- [Use Cases & Features by Category](#use-cases--features-by-category)
  - [1. Local Network Sharing & Collaboration](#1-local-network-sharing--collaboration)
    - [Live Project Directory Hosting (`cast live <path>`)](#live-project-directory-hosting-cast-live-path)
    - [Dev Server Reverse Proxying (`cast live <port>`)](#dev-server-reverse-proxying-cast-live-port)
    - [Connecting to Live Sessions (`cast connect <code>`)](#connecting-to-live-sessions-cast-connect-code)
    - [File & Folder Sharing (`cast share <path>`)](#file--folder-sharing-cast-share-path)
    - [Receiving Shared Files (`cast receive <code>`)](#receiving-shared-files-cast-receive-code)
  - [2. Local Clipboard Management](#2-local-clipboard-management)
    - [Interactive Clipboard Picker (`cast clip`)](#interactive-clipboard-picker-cast-clip)
    - [Clearing Clipboard History (`cast clip clear`)](#clearing-clipboard-history-cast-clip-clear)
  - [3. Safe Shell Arithmetic & Calculation](#3-safe-shell-arithmetic--calculation)
    - [Evaluating Math Expressions (`cast calc`)](#evaluating-math-expressions-cast-calc)
  - [4. Fast & Interactive Unit Conversion](#4-fast--interactive-unit-conversion)
    - [Interactive & Direct Conversion (`cast conv`)](#interactive--direct-conversion-cast-conv)
    - [Supported Units Reference](#supported-units-reference)
- [Networking & Discovery Architecture](#networking--discovery-architecture)
- [Configuration & Environment](#configuration--environment)
- [Security & Privacy Guarantees](#security--privacy-guarantees)
- [Uninstallation & Cleanup](#uninstallation--cleanup)
- [Development & Testing](#development--testing)
- [License](#license)

---

## Core Principles

- 🔒 **Local-First & Private**: No cloud accounts, external relays, or tracking telemetry. All traffic stays strictly on your local network or loopback interface.
- ⚡ **Zero Setup & Instant Sharing**: Share dev servers, static directories, and files across computers and mobile devices with human-friendly 6-character codes or terminal QR codes.
- 🛡️ **Defensive & Safe**: Built-in path traversal guards, non-destructive file downloads (never overwrites existing files), and automatic secret filtering for clipboard history.
- 🧰 **Developer Utility Belt**: Offline arithmetic evaluation and unit conversions without leaving the terminal or spawning heavy runtime interpreters.

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
# Binary is generated at ./cast
```

### Shell Completion

Generate command autocompletion scripts:

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
| `cast connect` | `<code>` | Discover and connect to an active live session by short code |
| `cast share` | `<file>` or `<dir>` | Share a file or folder on the local network (10-minute expiry) |
| `cast receive` | `<code>` | Download a shared file or directory archive by short code |
| `cast clip` | _none_ | Launch interactive TUI clipboard history picker |
| `cast clip clear` | _none_ | Delete all stored local clipboard history |
| `cast calc` | `<expression>` | Evaluate arithmetic expressions without shell execution |
| `cast conv` | `<val> <unit> [target]`| Convert between units (interactive picker if target omitted) |
| `cast completion` | `<shell>` | Generate autocompletion script (`bash`, `zsh`, `fish`, `powershell`) |
| `cast version` | _none_ | Print current `cast` version |

**Global Flags:**
- `--debug`: Show diagnostic logging on `stderr`.
- `-h`, `--help`: Display help information for any command.
- `-v`, `--version`: Print `cast` version.

---

## Use Cases & Features by Category

### 1. Local Network Sharing & Collaboration

#### Live Project Directory Hosting (`cast live <path>`)

Turn any folder into an instant local HTTP file server accessible from phones, tablets, or other computers on the same Wi-Fi.

```bash
# Serve the current directory
cast live .

# Serve a specific project directory
cast live ./my-static-site
```

**Features:**
- **Interactive Web Interface**: Serving a directory provides a clean HTML web directory browser with parent navigation (`..`) and clickable links for browsing subfolders and viewing/downloading files.
- **MIME Detection**: Automatically sets correct `Content-Type` headers based on file extension.
- **Terminal QR Code**: Renders a compact half-block ASCII QR code for instant mobile camera scanning.
- **Dual URL Output**: Displays both `Local URL` (`http://localhost:...`) and `Remote URL` (`http://<LAN_IP>:...`) with 32-byte URL-safe authorization tokens.

#### Dev Server Reverse Proxying (`cast live <port>`)

Test web apps on mobile devices or share running local development servers without deploying to staging or configuring complex tunnels.

```bash
# Reverse-proxy a local service running on port 3000 (e.g. Next.js, Vite, Rails)
cast live 3000

# Proxy backend APIs or services on another port
cast live 8080
```

**Features:**
- **Reverse Proxy**: Transparently proxies HTTP requests to your local dev service while enforcing token-based authorization.
- **Path & Query Preservation**: Correctly handles subpaths, route rewriting, and URL parameters.
- **Graceful Error Reporting**: Detects whether the local service is reachable before initiating proxying.

#### Connecting to Live Sessions (`cast connect <code>`)

Connect to an active live session from another computer on the same local network:

```bash
cast connect 7F92K8
```

- Discovers the session via local UDP broadcast (`port 39422`).
- Displays the hosting device's name (from cryptographic device identity or OS hostname), project path, and remote URL.

#### File & Folder Sharing (`cast share <path>`)

Transfer files and folders directly across machines at local network speeds without third-party cloud uploads.

```bash
# Share a single document or media file
cast share presentation.pdf

# Share an entire directory
cast share ./project-assets
```

**Features:**
- **On-the-Fly Directory Compression**: Folders are automatically compressed using `zip.Deflate` into temporary archives that are cleaned up upon server termination or expiration.
- **Direct Web Downloads**: Recipients can download either via the CLI (`cast receive`) or directly in any browser by scanning the terminal QR code.
- **Content Headers**: Sets proper `Content-Disposition`, `Content-Length`, and MIME types.

#### Receiving Shared Files (`cast receive <code>`)

Download files shared with `cast share` using the 6-character code:

```bash
cast receive 7F92K8
```

**Features:**
- **Live Transfer Progress**: Displays animated transfer percentage (`XX%`), byte counters, and total file size.
- **Collision Protection**: Strictly fails safely if a file with the same name already exists in the destination directory to prevent accidental data loss.

---

### 2. Local Clipboard Management

#### Interactive Clipboard Picker (`cast clip`)

A private, lightweight clipboard history manager built into your terminal.

```bash
cast clip
```

**Features:**
- **Interactive TUI**: Powered by Bubble Tea with intuitive controls:
  - `↑` / `↓` or Vim keys `k` / `j` to navigate entries.
  - Number keys `1`..`9` to jump directly and copy an item.
  - `Enter` to restore the selected text back into your system clipboard.
  - `q`, `Esc`, or `Ctrl+C` to exit without changing clipboard contents.
- **Automatic Deduplication**: Consecutive identical entries are never stored twice.
- **Background Watcher**: Monitors clipboard changes in the background while `cast clip` is open.
- **Sensitive Data Redaction**: Automatically filters out and refuses to record sensitive text (such as `begin private key`, `password=`, `api_key=`, `authorization: bearer`, or `bearer `).
- **SQLite Persistence**: Stored in a local SQLite database (`history.db`) with owner-only access permissions and configurable retention limit.
- **Platform Support**: Native clipboard integration on macOS (`pbcopy`/`pbpaste`).

#### Clearing Clipboard History (`cast clip clear`)

Permanently wipe all locally stored clipboard history:

```bash
cast clip clear
```

---

### 3. Safe Shell Arithmetic & Calculation

#### Evaluating Math Expressions (`cast calc`)

Perform decimal calculations directly in your terminal without invoking external scripting engines or risking shell injection.

```bash
# Basic arithmetic
cast calc 15 * 32
cast calc 100 / 4 + 2.5

# Parentheses and order of operations
cast calc '(12 + 8) * 3'

# Exponentiation and modulo
cast calc 2 ^ 10
cast calc 27 % 4

# Unary operators
cast calc -5 * (10 + -2)
```

**Features:**
- **Operators Supported**: Addition (`+`), subtraction (`-`), multiplication (`*`), division (`/`), modulo (`%`), power (`^`), unary minus/plus (`-`/`+`), parentheses (`(...)`).
- **Safety**: Evaluated using a custom recursive-descent parser. No shell `eval`, no Python/Node subprocesses, zero code injection vulnerabilities.
- **Shell Friendly**: Pass expressions as separate arguments or as a quoted string.

---

### 4. Fast & Interactive Unit Conversion

#### Interactive & Direct Conversion (`cast conv`)

Convert measurements across length, mass, temperature, time, data size, and speed.

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

**Features:**
- **Interactive Target Selector**: When target unit is omitted, displays all valid units in that category with arrow keys, Vim keys (`j`/`k`), and number shortcuts.
- **Precise Conversions**: Exact base conversions implemented offline.

#### Supported Units Reference

All unit conversions run 100% offline. Data units follow standard decimal (base-10) definitions:

| Category | Symbols | Full Names |
| :--- | :--- | :--- |
| **Length** | `mm`, `cm`, `m`, `km`, `in`, `ft`, `yd`, `mi` | Millimeters, centimeters, meters, kilometers, inches, feet, yards, miles |
| **Mass** | `mg`, `g`, `kg`, `oz`, `lb` | Milligrams, grams, kilograms, ounces, pounds |
| **Temperature** | `C`, `F`, `K` | Celsius, Fahrenheit, Kelvin |
| **Time** | `ms`, `s`, `min`, `h`, `day` | Milliseconds, seconds, minutes, hours, days |
| **Data Size** | `B`, `KB`, `MB`, `GB`, `TB` | Bytes, kilobytes (10³), megabytes (10⁶), gigabytes (10⁹), terabytes (10¹²) |
| **Speed** | `m/s`, `km/h`, `mph`, `ft/s` | Meters/sec, km/h, miles per hour, feet/sec |

---

## Networking & Discovery Architecture

`cast` uses a lightweight, zero-configuration local discovery system designed for LAN environments:

- **UDP Broadcast Ports**:
  - `39421`: Discovery responder for file shares (`cast share` & `cast receive`).
  - `39422`: Discovery responder for live sessions (`cast live` & `cast connect`).
- **Dual-Interface Probing**: Queries are simultaneously broadcast to both `255.255.255.255` (LAN broadcast) and `127.0.0.1` (loopback). This allows two `cast` instances on the same machine to seamlessly discover each other during testing or local workflows.
- **Human-Friendly Short Codes**: Codes use an unambiguous Base-32 alphabet (`23456789ABCDEFGHJKLMNPQRSTUVWXYZ`), purposefully excluding confusing characters (`0`, `O`, `1`, `I`).
- **Session Tokens**: Every session generates an independent 32-byte URL-safe cryptographic secret. HTTP endpoints verify tokens using constant-time comparison (`subtle.ConstantTimeCompare`) to resist timing attacks on shared networks.

---

## Configuration & Environment

`cast` runs out of the box with sensible defaults and requires no configuration files. Optional custom settings can be specified in your system configuration directory:

- **macOS:** `~/Library/Application Support/cast/config.toml` (or `~/.config/cast/config.toml`)
- **Linux:** `~/.config/cast/config.toml`
- **Windows:** `%APPDATA%\cast\config.toml`

### Example `config.toml`

```toml
[clipboard]
max_history = 200      # Number of clipboard items to retain in SQLite

[share]
expires_minutes = 10   # Lifetime for share and live sessions

[network]
preferred_port = 0     # 0 selects a random available OS port
```

### Environment Variables

| Variable | Description |
| :--- | :--- |
| `CAST_CONFIG_DIR` | Override the default directory for configuration, SQLite database (`history.db`), and device identity keys (`device.json`). Useful for testing and sandbox isolation. |

---

## Security & Privacy Guarantees

- **No Data Telemetry**: `cast` collects zero analytics, error telemetry, or diagnostic tracking.
- **Cryptographic Device Identity**: Generates an Ed25519 keypair and persistent device ID (`device.json`) on first run with owner-only file permissions (`0600`).
- **Path Traversal Protection**: File and directory servers strictly validate that requested file paths cannot escape the designated project root directory, resolving symlinks defensively.
- **Sensitive Content Shield**: Clipboard history automatically detects and refuses to persist credentials, authorization headers, private keys, and passwords.
- **Safe Overwrite Policy**: `cast receive` will never overwrite an existing file silently.

---

## Uninstallation & Cleanup

To completely remove `cast` and all associated data from your system:

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

### 2. Remove Local Configuration & Data

To delete clipboard history (`history.db`), cryptographic identity keys (`device.json`), and configuration:

```bash
# macOS
rm -rf ~/Library/Application\ Support/cast ~/.config/cast

# Linux
rm -rf ~/.config/cast

# Windows (PowerShell)
Remove-Item -Recurse -Force "$env:APPDATA\cast"
```

---

## Development & Testing

```bash
# Format source code
make fmt

# Run static analysis
make vet

# Run all unit and integration tests
make test

# Compile binary
make build
```

---

## License

This project is licensed under the MIT License.
