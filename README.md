# cast

`cast` is a fast, local-first command-line utility for developers and power users. It simplifies sharing files and local projects across devices on your local network, stores clipboard history, evaluates arithmetic expressions, and converts units with an interactive selector.

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

---

## Capabilities & Commands

### 1. Live Projects & Services

Expose a local directory or reverse-proxy a running local HTTP service over the local network with tokenized authorization and a terminal QR code:

```bash
# Serve current directory
cast live .

# Serve a specific project directory
cast live ./my-project

# Proxy a local service running on port 3000
cast live 3000
```

Connect to an active live session from another terminal or device:

```bash
cast connect 7F92K8
```

### 2. File & Directory Sharing

Share a single file or an entire directory. Folders are automatically compressed into temporary streaming ZIP archives and cleaned up upon expiration or server exit:

```bash
# Share a file
cast share report.pdf

# Share an entire directory (streamed as a temporary ZIP archive)
cast share ./my-folder
```

Receive a shared file or directory using its short code without overwriting existing files:

```bash
cast receive 7F92K8
```

### 3. Clipboard History

Browse and restore local clipboard history using SQLite and an interactive Bubble Tea terminal picker:

```bash
# Open interactive history picker
cast clip

# Clear clipboard history
cast clip clear
```

### 4. Calculator

Evaluate mathematical expressions safely using a dedicated recursive-descent/shunting-yard parser with zero shell execution:

```bash
cast calc 15 * 32
cast calc '(12 + 8) * 3'
cast calc 100 / 4 + 2.5
```

### 5. Unit Conversion

Convert values across length, mass, time, temperature, data size, and speed categories with interactive Bubble Tea selection or direct target arguments:

```bash
# Interactive menu
cast conv 10 km

# Direct conversion
cast conv 10 km mi
cast conv 100 C F
cast conv 2.5 GB MB
```

---

## Configuration

`cast` runs out of the box with sensible defaults and requires no configuration files or accounts. Optional configuration can be placed in `~/.config/cast/config.toml`:

```toml
[clipboard]
max_history = 200

[share]
expires_minutes = 10

[network]
preferred_port = 0
```

Set the `CAST_CONFIG_DIR` environment variable to isolate configuration during testing or local development.

---

## Security & Privacy

- **Local-first**: Operates entirely offline and on local networks. No accounts, telemetry, or remote analytics.
- **Tokenized authorization**: Displayed 6-character codes are human-friendly identifiers; network HTTP requests require high-entropy random secrets encoded into QR codes and URLs.
- **Path traversal and symlink guards**: Strict root containment ensures directory hosting never escapes the intended root directory.
- **Safe receives**: `cast receive` never silently overwrites an existing file.
- **No data logging**: Clipboard history, shared contents, private keys, and authorization tokens are never logged.
- **Cryptographic Device Identity**: Local Ed25519 keypair and device fingerprint are stored locally in the user configuration directory with restricted permissions (`0o600`).

---

## Development

```bash
# Format code
make fmt

# Run unit and integration tests
make test

# Run code vetting
make vet

# Build binary
make build
```
