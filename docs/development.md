# Development

## Running tests

```bash
go test ./...
```

With verbose output:

```bash
go test -v ./...
```

---

## Cross-compilation

Build for Windows from any platform:

```bash
GOOS=windows GOARCH=amd64 go build -o screensaver.exe ./cmd/screensaver
```
