# Build Instructions

## Prerequisites
- **Build Tool**: Go 1.22+ (`go version`)
- **Dependencies**: Managed via `go.mod` / `go.sum` (auto-downloaded)
- **Environment Variables**: None required
- **System Requirements**: Any POSIX-compatible OS (Linux, macOS, BSD), terminal with mouse support

## Build Steps

### 1. Install Dependencies
```bash
go mod download
```

### 2. Build the Binary
```bash
go build -o pidshooter .
```

### 3. Build for Multiple Platforms (optional)
```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o pidshooter-linux-amd64 .

# macOS (arm64 - Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o pidshooter-darwin-arm64 .

# macOS (amd64 - Intel)
GOOS=darwin GOARCH=amd64 go build -o pidshooter-darwin-amd64 .

# FreeBSD
GOOS=freebsd GOARCH=amd64 go build -o pidshooter-freebsd-amd64 .
```

### 4. Verify Build Success
- **Expected Output**: A single `pidshooter` binary in the project root
- **Verify**: `./pidshooter --help` should print usage information
- **Binary size**: ~5-8 MB (statically linked Go binary)

### 5. Install (optional)
```bash
go install .
# Binary will be placed in $GOPATH/bin or $HOME/go/bin
```

## Static Analysis

### Run go vet
```bash
go vet ./...
```

### Run staticcheck (optional, install first)
```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

## Troubleshooting

### Build Fails with "module not found"
- **Cause**: Dependencies not downloaded
- **Solution**: Run `go mod download` or `go mod tidy`

### Build Fails with "go version too old"
- **Cause**: Go version < 1.22
- **Solution**: Update Go to 1.22+: https://go.dev/dl/

### Binary runs but no mouse support
- **Cause**: Terminal doesn't support xterm mouse protocol
- **Solution**: Use a modern terminal (iTerm2, Alacritty, GNOME Terminal, kitty, etc.)
