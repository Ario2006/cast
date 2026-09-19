# CAST — Prototype Application Specification

> Build-ready specification for the `cast` CLI.
>
> Goal: create a useful, cross-platform developer utility that is simple to understand, easy to run, and structured so the prototype can evolve into a serious long-term product without a rewrite.

---

## 1. Product Overview

`cast` is a local-first command-line utility for developers and power users.

The prototype has five primary capabilities:

1. **Live a project or local service** and access it from another device.
2. **Share a file/folder using a short code or QR code.**
3. **Store and browse clipboard history.**
4. **Evaluate calculations from the command line.**
5. **Convert units through an interactive selection menu.**

The product should feel fast and frictionless:

- No account is required for the prototype.
- Most functionality should work locally/offline.
- Network functionality should use the local network first.
- Commands should be memorable and predictable.
- Interactive menus should be optional where practical, but should make common tasks pleasant.
- Avoid unnecessary abstraction and complicated distributed-systems logic in the prototype.

The project name is **cast**.

---

## 2. Core Product Principles

### 2.1 Simple first

Prefer the simplest implementation that works reliably.

Do not add:

- microservices
- cloud infrastructure
- plugin systems
- accounts/authentication servers
- complex sync/conflict engines
- custom binary protocols where a simple encoding works
- unnecessary dependency layers

unless they are explicitly needed for the current prototype.

### 2.2 Modular internals

Keep features separated behind small interfaces so the prototype can evolve later.

Good:

```text
CLI command
    -> application/service
        -> small interface
            -> concrete implementation
```

Avoid:

```text
CLI command
    -> networking
    -> database
    -> filesystem
    -> terminal UI
    -> crypto
    -> protocol
```

all directly in one command handler.

### 2.3 Local-first

`cast` should not require a cloud backend for the MVP.

For network features, prefer:

```text
same machine
   -> LAN
   -> direct remote connection later
   -> optional relay later
```

### 2.4 Human-readable behavior

Output should be understandable without reading documentation.

Example:

```text
$ cast share ./report.pdf

Sharing: report.pdf
Code:    7F92K
URL:     http://192.168.1.20:8765/s/7F92K

Scan the QR code or run:
  cast receive 7F92K

Expires in 10 minutes.
```

### 2.5 Good errors

Errors should explain:

1. what failed
2. why it likely failed
3. what the user can do next

Example:

```text
Could not connect to 7F92K.
The code may have expired or the sharing device may be offline.
```

---

# 3. Technology Stack

## 3.1 Language

**Go**

Use a current stable Go release available when implementation starts.

Why:

- fast development
- excellent networking support
- simple concurrency
- straightforward filesystem APIs
- easy cross-platform compilation
- single-binary distribution
- strong long-term fit for a systems/networking CLI

## 3.2 CLI framework

**Cobra**

Use Cobra for commands, flags, arguments, help text, and command hierarchy.

Primary commands:

```text
cast live
cast connect
cast share
cast receive
cast clip
cast calc
cast conv
```

Optional future commands:

```text
cast device
cast doctor
cast version
```

## 3.3 Interactive terminal UI

**Bubble Tea**

Use it only for genuinely interactive screens, especially:

- clipboard history selection
- conversion target selection
- future device/session selection

Do not turn every command into a TUI.

## 3.4 Database

**SQLite**

Use SQLite for local metadata and history.

Suggested library approach:

- SQLite driver compatible with Go
- `sqlc` for typed SQL if practical

Do not use an ORM unless it clearly reduces complexity.

SQLite stores metadata, not large shared files.

## 3.5 Networking

Prototype:

- HTTP for browser-accessible `cast live` pages
- simple TCP/TLS or HTTP-based transfer for device-to-device operations where practical

Later:

- QUIC for long-lived/persistent sessions
- NAT traversal / relay support

Do not start by implementing full peer-to-peer traversal.

## 3.6 Service discovery

Prototype:

- direct local IP + port in generated URL
- QR code containing the connection URL/payload
- short code for the `cast connect` / `cast receive` flow

Later:

- mDNS for LAN discovery

## 3.7 Serialization

For the prototype, use a simple human-readable format or small JSON payloads where practical.

Do not introduce Protobuf until the session protocol has enough complexity to justify it.

Design the protocol layer so Protobuf can be introduced later without changing application logic.

## 3.8 Cryptography

Use Go's standard cryptography/TLS packages where possible.

Do not implement cryptographic primitives manually.

The prototype should at minimum avoid transmitting shared content in plaintext across networks that are not trusted.

## 3.9 Filesystem watching

Use **fsnotify** for project change detection.

## 3.10 QR generation

Use a small, well-maintained QR-code library that can render a terminal-friendly QR code.

The QR should contain a URL or compact connection payload, not the file contents themselves.

## 3.11 Build and releases

Use **GoReleaser** for packaging releases later.

Target:

- macOS
- Linux
- Windows

The MVP should run on the primary development platform first, while avoiding OS-specific assumptions in core code.

## 3.12 CI

Use GitHub Actions for:

- formatting check
- unit tests
- build
- cross-platform compile checks

---

# 4. Supported Platforms

Primary target:

- macOS

Also target:

- Linux
- Windows

The architecture must avoid making the core application dependent on one operating system.

OS-specific code should live under a small platform package.

Example:

```text
internal/platform/clipboard
    clipboard_darwin.go
    clipboard_linux.go
    clipboard_windows.go
```

Do not spread OS-specific calls throughout the application.

---

# 5. Command-Line Interface

## 5.1 General command rules

Global behavior:

```text
cast --help
cast --version
```

Use standard CLI conventions:

- exit code `0` for success
- non-zero for failures
- useful stderr for errors
- stdout for normal command output

Do not require quotes for calculations when the shell does not require them.

If shell expansion or special characters can alter the expression, quotes may still be used normally.

---

# 6. `cast live`

## 6.1 Purpose

Make a local project directory or local web application/service accessible from another device.

## 6.2 Syntax

```bash
cast live .
cast live ./project
cast live 3000
cast live ./project 3000
```

Supported interpretation:

- `.` = current directory
- path = project directory
- integer port = local service port
- path + port = serve/access the project while exposing the specified local service

The exact handling of path + port must be clearly documented in CLI help.

For the prototype, the simplest implementation is:

### Directory mode

```bash
cast live ./my-project
```

Host the directory over HTTP for browsing/download.

### Port mode

```bash
cast live 3000
```

Proxy the existing local service on port `3000`.

### Current-directory shortcut

```bash
cast live .
```

Equivalent to using the current working directory.

## 6.3 Output

On success, show:

```text
Casting live:

Project: ./my-project

Local URL:
  http://localhost:8765

Remote URL:
  http://192.168.1.20:8765

Live code:
  7F92K8

Scan the QR code or run:
  cast connect 7F92K8

Press Ctrl+C to stop.
```

The exact URL/port can differ, but the output structure should remain similar.

## 6.4 QR behavior

The QR should encode enough information for another device to connect.

Prefer a URL such as:

```text
http://192.168.1.20:8765/c/7F92K8
```

Do not encode the project files in the QR.

## 6.5 Short code

Generate a human-friendly code.

Requirements:

- 4–8 characters preferred
- uppercase letters and digits
- avoid ambiguous characters where possible (`O`, `0`, `I`, `1`)
- short-lived by default
- unique among currently active local sessions

Recommended display format:

```text
7F92K8
```

The implementation may use a random 6–8 character token internally.

## 6.6 Connection flow

Another device can:

1. open the URL
2. scan the QR
3. run:

```bash
cast connect 7F92K8
```

For the MVP, `cast connect` should first attempt to resolve the code on the local network/session registry.

If the connection code cannot be resolved, return a clear error.

## 6.7 MVP scope

The first `live` version should support:

- project directory hosting
- existing local HTTP service proxying
- active session code
- QR
- remote URL
- graceful shutdown

Do not implement project synchronization in the first `live` version.

"Live" means **access the current project/service remotely**, not bidirectional filesystem sync yet.

Future versions can add live file synchronization.

---

# 7. `cast connect`

## 7.1 Purpose

Connect to a live session created by `cast live`.

## 7.2 Syntax

```bash
cast connect 7F92K8
```

## 7.3 Behavior

For a directory-hosting live session:

- connect to the remote device
- display/open the available remote URL
- optionally provide a terminal-readable file listing

For a local web-service live session:

- open or print the proxied URL

For the MVP, connecting can simply establish/resolve the session and show the remote URL.

Later, this command can become a richer remote session command.

## 7.4 Output

Example:

```text
Connected to:
  Aryan's MacBook

Live project:
  website

URL:
  http://192.168.1.20:8765
```

Avoid printing unnecessary technical details unless `--verbose` is used.

---

# 8. `cast share`

## 8.1 Purpose

Share a file or folder with another device.

Examples:

```bash
cast share image.png
cast share video.mp4
cast share report.pdf
cast share notes.txt
cast share ./folder
```

Supported inputs:

- images
- videos
- audio
- text files
- PDFs
- archives
- arbitrary files
- directories

For directories, create a temporary ZIP archive for sharing.

Do not permanently modify the original directory.

## 8.2 Output

Example:

```text
Sharing:
  report.pdf

Size:
  8.4 MB

Code:
  7F92K8

Scan the QR code to download.

Or run:
  cast receive 7F92K8

Expires in:
  10 minutes
```

Display a terminal QR code.

## 8.3 Share URL

The QR should resolve to a download endpoint similar to:

```text
http://192.168.1.20:8765/s/7F92K8
```

The endpoint should present a simple download page or begin the download directly.

## 8.4 Share lifecycle

A share should have:

```text
id
source path
created time
expiration time
status
share type
```

Default expiration:

```text
10 minutes
```

The exact default can be configured later.

## 8.5 One-time behavior

The default prototype behavior may allow multiple downloads until expiration.

Optional future flag:

```bash
cast share file.pdf --once
```

Do not make one-time behavior a requirement for MVP unless trivial to implement.

## 8.6 Cleanup

When a share expires:

- invalidate the code
- stop serving it
- delete any temporary ZIP archive created for a folder

Do not delete the user's original files.

---

# 9. `cast receive`

## 9.1 Purpose

Download a shared file using its code.

## 9.2 Syntax

```bash
cast receive 7F92K8
```

## 9.3 Behavior

The command should:

1. resolve the active share
2. connect to the sharing device
3. receive the file
4. write it to the current directory by default
5. show progress for larger files

Example:

```text
Receiving report.pdf...

██████████████████░░░  82%
8.1 MB / 9.8 MB
```

## 9.4 Destination

Optional future support:

```bash
cast receive 7F92K8 --output ./downloads
```

For MVP, current directory is enough unless implementation is trivial.

## 9.5 Security

Do not silently overwrite an existing file.

If the destination exists:

```text
report.pdf already exists.
Choose:
  [1] rename
  [2] overwrite
  [3] cancel
```

For non-interactive use, fail safely rather than overwrite.

---

# 10. `cast clip`

## 10.1 Purpose

Store and browse local clipboard history.

## 10.2 Basic command

```bash
cast clip
```

This opens an interactive clipboard-history picker.

## 10.3 UI requirements

Display recent items in reverse chronological order.

Example:

```text
Clipboard History

> 1  npm install drizzle-orm
  2  https://github.com/example/repo
  3  SELECT * FROM users WHERE id = ?
  4  192.168.1.42
  5  Meeting at 4:30 PM

↑ ↓ move   Enter select   q quit
```

User interactions:

- arrow up/down = move selection
- Enter = choose selected item
- number key = immediately choose item by number
- `q` / Escape = exit without changing clipboard

After selection:

- put the selected item back into the OS clipboard
- exit successfully

## 10.4 Supported clipboard content

MVP:

- plain text

Future:

- images
- rich text
- files

Only support text history in the prototype unless native image clipboard support is already straightforward on the target OS.

## 10.5 Clipboard watcher

The clipboard service should monitor changes in the background while `cast clip` is running.

History is stored locally in SQLite.

Do not require a permanent system daemon for MVP.

Later, a background daemon can maintain history continuously.

## 10.6 Limits

To prevent unlimited storage:

- keep a configurable history size
- default to a reasonable number such as 100–500 entries
- store timestamps
- deduplicate consecutive identical clipboard entries

## 10.7 Sensitive content

The implementation should avoid obvious accidental storage of secrets where practical.

For MVP:

- provide a clear local-only behavior
- never sync clipboard history to another device
- allow users to clear history

Future:

```bash
cast clip clear
cast clip pause
cast clip exclude
```

and optional secret detection/redaction.

---

# 11. `cast calc`

## 11.1 Purpose

Evaluate mathematical expressions.

Syntax:

```bash
cast calc 15 * 32
cast calc 100 / 4
cast calc "15 * 32"
cast calc '(12 + 8) * 3'
```

The parser should behave consistently regardless of normal whitespace differences.

Quotes should be optional where shell parsing permits the expression.

## 11.2 Required operators

MVP:

- `+`
- `-`
- `*`
- `/`
- `%`
- parentheses
- decimal numbers

Recommended additional operators when easy to implement:

- exponentiation
- unary minus

## 11.3 Safety

Do not use shell execution, `eval`, or arbitrary code execution.

The expression must be parsed by a dedicated calculator parser/evaluator.

## 11.4 Output

Example:

```text
480
```

Prefer concise output for simple calculations.

For invalid expressions:

```text
Could not evaluate expression.
Check the syntax around: `* 32`
```

## 11.5 Parsing strategy

Use a small expression parser with a clear grammar, such as recursive descent or shunting-yard.

Do not depend on a large general-purpose scripting engine.

---

# 12. `cast conv`

## 12.1 Purpose

Convert a value from one unit to another using an interactive target-unit selection menu.

Syntax:

```bash
cast conv 10 km
cast conv 72 mph
cast conv 5 ft
cast conv 100 C
cast conv 2.5 GB
```

The command may optionally support a direct target in the future:

```bash
cast conv 10 km mi
```

But the MVP should focus on the interactive menu.

## 12.2 Behavior

1. Parse the input number and unit.
2. Determine valid target units.
3. Show an interactive menu.
4. Allow selection with arrow keys + Enter.
5. Allow selection by number.
6. Display the converted result.

Example:

```text
10 km →

> 1. meters (m)
  2. miles (mi)
  3. feet (ft)
  4. yards (yd)

↑ ↓ move   Enter select
```

Alternatively, pressing `2` should immediately select miles.

## 12.3 Unit categories

MVP categories:

### Length

- mm
- cm
- m
- km
- in
- ft
- yd
- mi

### Mass

- mg
- g
- kg
- oz
- lb

### Temperature

- C
- F
- K

### Time

- ms
- s
- min
- h
- day

### Data size

- B
- KB
- MB
- GB
- TB

Choose and document whether data units use decimal or binary definitions. Prefer decimal units for the first version unless implementation constraints suggest otherwise.

### Speed

- m/s
- km/h
- mph
- ft/s

The unit engine should be designed so more categories can be added without changing command code.

## 12.4 Currency

Currency conversion may be included as a separate category if a reliable live exchange-rate provider is configured.

Do not pretend currency rates are offline/static.

Use a provider interface:

```go
type CurrencyProvider interface {
    Rate(ctx context.Context, from, to string) (float64, error)
}
```

If live currency data is unavailable, return a clear error rather than a fake or stale value.

---

# 13. Future inline mode

Do not make this an MVP requirement, but design the command parser so a future shortcut can be added:

```bash
cast "15 * 32"
cast "10 km in mi"
```

For now, explicit `calc` and `conv` commands are sufficient.

---

# 14. High-Level Architecture

Use a simple layered architecture.

```text
                    +----------------------+
                    |       `cast` CLI     |
                    |  Cobra + Bubble Tea  |
                    +----------+-----------+
                               |
                    +----------v-----------+
                    |   Application Layer  |
                    |----------------------|
                    | LiveService          |
                    | ShareService         |
                    | ClipboardService     |
                    | CalculatorService    |
                    | ConversionService    |
                    +----------+-----------+
                               |
                    +----------v-----------+
                    |      Domain          |
                    |----------------------|
                    | Share                |
                    | Session              |
                    | Device               |
                    | ClipboardEntry       |
                    | Quantity             |
                    +----------+-----------+
                               |
              +----------------+----------------+
              |                |                |
      +-------v------+ +-------v------+ +-------v------+
      |   Storage    | |  Transport   | |  Platform    |
      |   SQLite     | | TCP/TLS now  | | Clipboard    |
      |              | | QUIC later   | | OS helpers   |
      +--------------+ +--------------+ +--------------+
```

The key rule is that **commands should call application services instead of implementing business logic themselves**.

---

# 15. Suggested Repository Structure

Use a structure that is easy to navigate.

```text
cast/
├── cmd/
│   └── cast/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── live/
│   │   ├── share/
│   │   ├── clipboard/
│   │   ├── calculator/
│   │   └── conversion/
│   │
│   ├── domain/
│   │   ├── session/
│   │   ├── share/
│   │   ├── device/
│   │   └── clipboard/
│   │
│   ├── transport/
│   │   ├── transport.go
│   │   └── tcp/
│   │
│   ├── storage/
│   │   ├── db.go
│   │   ├── migrations/
│   │   └── queries/
│   │
│   ├── platform/
│   │   └── clipboard/
│   │
│   ├── qr/
│   ├── units/
│   └── config/
│
├── migrations/
├── proto/                 # optional later
├── tests/
│   ├── integration/
│   └── e2e/
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── app-spec.md
```

Do not create packages just for the sake of having packages.

Keep related code together until the package has a real responsibility.

---

# 16. Application Interfaces

Use small interfaces where swapping implementations is likely.

## 16.1 Storage

Conceptually:

```go
type Storage interface {
    SaveShare(Share) error
    GetShare(id string) (Share, error)
    DeleteShare(id string) error

    AddClipboardEntry(ClipboardEntry) error
    ListClipboardEntries(limit int) ([]ClipboardEntry, error)
}
```

The exact interface can be adjusted during implementation.

## 16.2 Transport

Conceptually:

```go
type Transport interface {
    Listen(addr string) error
    Connect(ctx context.Context, addr string) (Connection, error)
}
```

Do not expose implementation-specific details to higher layers.

## 16.3 Clipboard

Conceptually:

```go
type Clipboard interface {
    Read() (string, error)
    Write(text string) error
}
```

The watcher can live as a separate service around this interface.

## 16.4 Currency provider

Conceptually:

```go
type CurrencyProvider interface {
    Rate(ctx context.Context, from, to string) (float64, error)
}
```

---

# 17. Device Identity

Create a simple local device identity early so the networking architecture can grow cleanly.

Suggested concept:

```text
Device
├── ID
├── name
└── key pair
```

Store the private key locally.

Example location:

```text
~/.config/cast/identity/
```

On first run:

```text
cast initialized.
Device: Aryan's MacBook
```

Do not build accounts or cloud authentication for the prototype.

---

# 18. Session Model

Both live and share operations should create a simple session.

Conceptual model:

```text
Session
├── ID / code
├── type
├── created_at
├── expires_at
├── device_id
├── address
└── permissions
```

Session types:

```text
live
share
```

Example permissions:

```text
read
write
execute (future)
```

For MVP:

- `share` = read/download only
- `live` directory mode = read/list/download only
- `live` service proxy = forward HTTP requests only

Do not enable arbitrary remote command execution in the prototype.

---

# 19. Networking Design for MVP

Keep the first network architecture simple.

## 19.1 Same-LAN assumption

The prototype should work when both devices are on the same local network.

Examples:

- laptop → phone
- laptop → another laptop
- desktop → laptop

## 19.2 Connection resolution

When a live/share session starts:

1. find local address(es)
2. start a listening server
3. generate a short session code
4. generate a QR payload
5. show connection instructions

For the first prototype, code resolution can be implemented with a small local mechanism rather than a cloud registry.

Possible MVP implementation:

- code identifies the session
- QR carries the actual connection URL
- `cast receive CODE` / `cast connect CODE` discovers the local active session through a simple local discovery mechanism or a local session endpoint

Avoid introducing remote code lookup infrastructure in MVP.

## 19.3 Later networking path

Design for this evolution:

```text
MVP:
HTTP/TCP + TLS
    ↓
LAN

Later:
QUIC/TLS
    ↓
LAN + direct P2P
    ↓
relay when direct P2P fails
```

Do not implement NAT traversal during the prototype unless required to demonstrate the product.

---

# 20. Live Directory Serving

For:

```bash
cast live .
```

serve files from the selected directory using a small HTTP server.

Requirements:

- directory listing
- file download
- sensible content types
- no path traversal
- restrict access to the shared project root
- do not serve parent directories
- stop server on Ctrl+C

Never allow `../` or equivalent traversal outside the selected root.

---

# 21. Live Port Proxying

For:

```bash
cast live 3000
```

proxy requests to the existing local service on port `3000`.

MVP scope:

- HTTP proxy
- request/response forwarding
- local target validation
- graceful shutdown

Do not support arbitrary raw TCP port forwarding initially.

Future versions can support broader protocols if needed.

---

# 22. File Sharing Implementation

For a file:

```text
original file
     ↓
share session
     ↓
HTTP GET /s/CODE
     ↓
download
```

For a directory:

```text
directory
     ↓
temporary ZIP
     ↓
share session
     ↓
download
```

Use streaming I/O for downloads instead of loading entire files into memory.

For large files:

- stream from disk
- show progress when receiving
- avoid full-file buffering

---

# 23. Database Model

A simple schema is enough.

## 23.1 `shares`

```text
id
code
path
name
type
created_at
expires_at
status
```

## 23.2 `clipboard_entries`

```text
id
content
created_at
```

## 23.3 `devices`

```text
id
name
public_key
created_at
last_seen_at
```

Only add more tables when the feature requires them.

---

# 24. Configuration

Suggested config path:

```text
~/.config/cast/config.toml
```

Mac/Linux should follow reasonable XDG conventions.

Windows should use the standard user configuration directory.

Possible settings:

```toml
[clipboard]
max_history = 200

[share]
expires_minutes = 10

[network]
preferred_port = 0
```

Do not require configuration for normal use.

Defaults should be sensible.

---

# 25. CLI UX Rules

## 25.1 Keep commands discoverable

```bash
cast --help
cast live --help
cast share --help
cast clip --help
cast calc --help
cast conv --help
```

## 25.2 Keep successful output short

The terminal should not be flooded with logs.

Use `--verbose` later for diagnostics.

## 25.3 Interactive vs non-interactive

Interactive commands:

- `cast clip`
- `cast conv ...`

Non-interactive commands:

- `cast calc ...`
- `cast share ...`
- `cast receive CODE`
- `cast live ...`

## 25.4 Ctrl+C

Ctrl+C must cleanly stop active servers and temporary resources.

Examples:

- stop live server
- invalidate active session
- remove temporary ZIP
- close database cleanly

---

# 26. Security Requirements

Security matters because this tool exposes files and local services.

## MVP mandatory requirements

1. Never expose files outside the explicitly selected root.
2. Prevent path traversal.
3. Never execute remote files automatically.
4. Never execute remote commands in MVP.
5. Use secure transport for network traffic where practical.
6. Use random, non-guessable session tokens underneath displayed codes.
7. Expire sessions.
8. Do not overwrite receiving files by default.
9. Do not store clipboard history remotely.
10. Never log clipboard contents.
11. Never log shared file contents.
12. Do not store private keys in the database.

## Important distinction

The displayed short code is a **human-friendly identifier**, not the sole cryptographic secret.

Internally, use a stronger random token for actual authorization.

Example:

```text
Displayed:
7F92K8

Internal token:
random high-entropy secret
```

The QR can carry the stronger token.

---

# 27. Privacy Requirements

By default:

- no user account
- no analytics
- no telemetry
- no cloud dependency
- clipboard stays local
- shared files are not uploaded to a central server

Do not add telemetry to the prototype unless explicitly requested later.

---

# 28. Logging

Normal user-facing commands should show friendly status messages.

Internal logs should go to a local log file or debug stream when enabled.

Never log:

- clipboard contents
- authentication tokens
- private keys
- full file contents

Allow a future `--verbose` or `cast doctor` workflow for debugging.

---

# 29. Error Handling

Define a small set of understandable error categories:

```text
invalid input
file not found
permission denied
session not found
session expired
connection failed
unsupported unit
invalid expression
port unavailable
```

Map them to useful user messages.

Avoid raw stack traces in normal CLI output.

---

# 30. Testing Strategy

The prototype must be functional, not just compilable.

## Unit tests

Test:

- code generation
- expiration logic
- calculator parser/evaluator
- unit conversion
- clipboard deduplication
- share metadata
- path validation

## Integration tests

Test real components together:

### File sharing

```text
server starts
    ↓
share file
    ↓
client downloads
    ↓
file checksum matches
```

### Live directory

```text
server starts
    ↓
HTTP GET file
    ↓
correct content returned
```

### Live port proxy

```text
local test server
    ↓
cast live PORT
    ↓
proxy request
    ↓
response matches
```

### Clipboard

Test service logic separately from OS clipboard behavior.

## End-to-end tests

At least one test should start two real `cast` processes and verify:

```text
share
    -> receive
    -> file exists
    -> checksum matches
```

---

# 31. Prototype Acceptance Criteria

The prototype is complete when all of the following work reliably.

## `cast live`

```bash
cast live .
```

must:

- start successfully
- show remote URL
- show short code
- print QR
- allow another device on the LAN to access the project
- stop cleanly with Ctrl+C

And:

```bash
cast live 3000
```

must:

- proxy an existing local HTTP application
- show remote URL
- show short code
- print QR

## `cast connect`

```bash
cast connect CODE
```

must:

- resolve an active local session
- report the connection
- show the useful remote URL
- clearly report expired/invalid codes

## `cast share`

```bash
cast share file.pdf
cast share image.png
cast share ./folder
```

must:

- support files
- support folders by creating a temporary ZIP
- show a short code
- show QR
- provide a receive command
- expire the share
- clean up temporary ZIP files

## `cast receive`

```bash
cast receive CODE
```

must:

- download the file
- show progress for larger files
- preserve filename
- avoid unsafe overwrite
- fail clearly for invalid/expired codes

## `cast clip`

```bash
cast clip
```

must:

- display recent clipboard text
- allow arrow navigation
- allow number-key selection
- copy selected value back to clipboard
- quit cleanly

## `cast calc`

```bash
cast calc 15 * 32
```

must return:

```text
480
```

It must support spaces, parentheses, and standard arithmetic operations without relying on shell execution.

## `cast conv`

```bash
cast conv 10 km
```

must:

- parse the value and unit
- show valid target units
- support arrow-key selection
- support number-key selection
- output the result

---

# 32. Explicit Non-Goals for the Prototype

Do NOT implement these unless needed for a demonstration:

- user accounts
- cloud dashboard
- centralized file storage
- social features
- sharing history across devices
- clipboard sync across devices
- remote command execution
- arbitrary TCP tunneling
- NAT traversal
- TURN infrastructure
- conflict-free distributed filesystem
- collaborative editing
- mobile application
- browser extension
- plugin marketplace
- analytics/telemetry

These may be considered later.

---

# 33. Long-Term Evolution Path

The prototype architecture should allow this eventual direction:

```text
                       cast
                        |
        +---------------+---------------+
        |               |               |
      Local           Device          Network
        |            Identity            |
        |               |               |
    SQLite          Sessions         Transport
        |               |               |
        |               |        +------+------+
        |               |        |             |
        |               |      Direct        Relay
        |               |        P2P           |
        |               |        |             |
        +---------------+--------+-------------+
                                |
                              QUIC
```

Potential future features:

- persistent device pairing
- mDNS discovery
- remote project sessions
- P2P connections across the internet
- relay fallback
- bidirectional project synchronization
- remote terminal sessions
- clipboard synchronization
- richer file transfers
- resumable downloads
- multiple simultaneous shares
- trusted device list
- mobile receiver
- web receiver
- cloud-assisted discovery without storing user files

The existing core concepts should remain:

```text
Device
Session
Transport
Share
Project
Clipboard
Calculator
Converter
```

---

# 34. Recommended Implementation Order

Implement in this exact general sequence unless a practical dependency requires adjustment.

## Step 1 — project foundation

Create:

- Go module
- Cobra root command
- version command
- config directory support
- basic error handling
- logging

## Step 2 — calculator

Implement:

- parser
- evaluator
- `cast calc`

This validates CLI argument handling.

## Step 3 — conversion engine

Implement:

- unit definitions
- conversion engine
- interactive selector
- `cast conv`

## Step 4 — clipboard

Implement:

- platform clipboard adapter
- SQLite schema
- history service
- Bubble Tea picker
- `cast clip`

## Step 5 — sharing

Implement:

- session/code generation
- local HTTP server
- single-file sharing
- QR generation
- `cast share`
- `cast receive`

## Step 6 — live directory

Implement:

- project root validation
- HTTP directory server
- QR
- live code
- remote URL
- `cast live .`

## Step 7 — live local port

Implement:

- HTTP reverse proxy
- `cast live PORT`

## Step 8 — connect

Implement:

- code resolution
- `cast connect CODE`

## Step 9 — identity and clean abstractions

Add:

- device identity
- transport interface
- session model
- platform isolation

Only do this once the basic product works.

## Step 10 — hardening

Add:

- more tests
- path traversal protection
- expiry cleanup
- better error messages
- cross-platform build verification
- packaging

---

# 35. Definition of Done for Code Quality

The implementation is acceptable when:

- `go test ./...` passes
- `go vet ./...` passes where applicable
- code is formatted with `gofmt`
- commands have useful `--help`
- core logic is unit tested
- at least one end-to-end sharing flow works
- the application can be built as one binary
- no secrets or clipboard contents are logged
- Ctrl+C leaves no obvious temporary resources behind
- the code can be understood by a developer unfamiliar with the project

Avoid cleverness.

Prefer:

```go
func CreateShare(...)
func ReceiveShare(...)
func Evaluate(...)
func Convert(...)
```

over deeply nested abstractions.

---

# 36. AI Agent Implementation Rules

The coding agent should follow these rules throughout development.

## 36.1 Build working slices

Prefer this:

```text
share a file end-to-end
```

before this:

```text
build a generalized networking framework
```

## 36.2 Keep dependencies minimal

Before adding a dependency, ask:

- Is it necessary?
- Does the standard library already solve this?
- Is the dependency maintained?
- Does it substantially simplify the code?

## 36.3 Do not prematurely optimize

Correctness and clarity come before:

- custom memory pooling
- custom binary protocols
- custom async frameworks
- complex caching

## 36.4 Keep network code isolated

Networking changes should mostly stay inside:

```text
internal/transport
internal/app/live
internal/app/share
```

## 36.5 Keep terminal UI isolated

Interactive UI code should not contain storage/networking logic.

## 36.6 Prefer explicit code

Good:

```text
CreateShare
GetShare
ExpireShare
ReceiveShare
```

Avoid generic framework code whose purpose is unclear.

## 36.7 No placeholder functionality presented as complete

If a feature is intentionally incomplete, clearly mark it as TODO and keep the currently supported path functional.

## 36.8 Preserve command behavior

Once a command works, avoid changing its user-facing syntax without a strong reason.

The intended primary CLI is:

```bash
cast live .
cast live 3000
cast connect CODE
cast share FILE
cast receive CODE
cast clip
cast calc EXPRESSION
cast conv VALUE UNIT
```

---

# 37. Suggested Initial README Examples

The README should eventually show the product through examples rather than lengthy explanations.

### Live a project

```bash
cast live .
```

Scan the QR code or run:

```bash
cast connect 7F92K8
```

### Share a file

```bash
cast share report.pdf
```

Scan the QR code or:

```bash
cast receive 7F92K8
```

### Clipboard

```bash
cast clip
```

Use arrow keys or the item number to restore an entry.

### Calculate

```bash
cast calc 15 * 32
```

### Convert

```bash
cast conv 10 km
```

Select the desired target unit.

---

# 38. Final Product Goal

`cast` should feel like a tiny utility that developers can install and immediately understand.

The desired mental model is:

```text
cast live     -> make something accessible
cast share    -> send something
cast receive  -> get something
cast connect  -> enter a live session
cast clip     -> retrieve something copied earlier
cast calc     -> calculate something
cast conv     -> convert something
```

The prototype should be deliberately small, reliable, and pleasant.

The long-term architecture should emerge from these working primitives rather than from premature infrastructure.

**Build the smallest version that demonstrates the product clearly, while preserving clean boundaries around sessions, transport, storage, platform integrations, and application logic.**
