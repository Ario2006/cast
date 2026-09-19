# cast

`cast` is a local-first developer CLI for sharing local files and projects, recalling clipboard entries, evaluating calculations, and converting units.

The prototype is under active development. The first available commands are:

```bash
cast --help
cast --version
cast version
cast calc '(12 + 8) * 3'
```

## Development

```bash
go test ./...
go vet ./...
go build ./cmd/cast
```

`cast` uses the platform configuration directory by default. Set `CAST_CONFIG_DIR` to isolate configuration during development or testing.
