# RDR: redis data reveal

RDR (redis data reveal) is a tool to parse Redis RDB files. Compared with [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools), RDR is implemented in Go and runs much faster (on my PC, a 5GB RDB file takes about 2 minutes).

## Build from Source (Windows / Linux)

Requires Go 1.21+ (recommended to use the toolchain 1.23.2 as declared in `go.mod`). On first build, you must generate embedded assets.

```bash
# 1) Install go-bindata (ensure the binary is in your PATH)
$ go install github.com/go-bindata/go-bindata/...@latest

# 2) Download dependencies
$ go mod tidy

# 3) Generate embedded static assets and templates (MUST do)
$ go generate ./...
```

### Local build on Linux

```bash
$ go build -o bin/rdr .
```

### Local build on Windows

```powershell
PS> go build -o bin/rdr.exe .
```

### Cross-compilation examples

```bash
# Build Windows binary from Linux
$ GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/rdr-windows-amd64.exe .

# Build Linux/amd64 and Linux/arm64 binaries from Linux
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/rdr-linux-amd64 .
$ GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/rdr-linux-arm64 .

# Build Linux binary from Windows (PowerShell)
PS> setx GOOS linux
PS> setx GOARCH amd64
PS> setx CGO_ENABLED 0
PS> go build -o bin/rdr-linux-amd64 .
```

Note: If you skip `go generate ./...`, the files `static/static.go` and `views/views.go` will be missing, causing `go build` to fail.

## Download

Latest binaries are attached to GitHub Releases. Use the following direct links to always fetch the latest version:

- Linux amd64: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-linux-amd64
- Linux arm64: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-linux-arm64
- macOS amd64: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-darwin-amd64
- macOS arm64 (Apple Silicon): https://github.com/ericdong2012/rdr/releases/latest/download/rdr-darwin-arm64
- Windows amd64 (.exe): https://github.com/ericdong2012/rdr/releases/latest/download/rdr-windows-amd64.exe

Checksums: https://github.com/ericdong2012/rdr/releases/latest/download/checksums.sha256

## Usage

```text
NAME:
   rdr - a tool to parse redis rdbfile

USAGE:
   rdr [global options] command [command options] [arguments...]

VERSION:
   v0.0.1

COMMANDS:
     dump     dump statistical information of rdbfile to STDOUT
     show     show statistical information of rdbfile by webpage
     export   export statistical information of rdbfile to local JSON or HTML
     keys     get all keys from rdbfile
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### show: view statistics on a web page

```text
NAME:
   rdr show - show statistical information of rdbfile by webpage

USAGE:
   rdr show [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --port value, -p value  Port for rdr to listen (default: 8080)
```

Example:

```bash
$ ./rdr show -p 8080 *.rdb
```

Note: memory usage is approximate.

![show example](https://yqfile.alicdn.com/img_9bc93fc3a6b976fdf862c8314e34f454.png)

### export: export statistics to local JSON or HTML

```text
NAME:
   rdr export - export statistical information of rdbfile to local JSON or HTML

USAGE:
   rdr export [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --format value, -f value  export format: json or html (default: "json")
   --out value, -o value     output file path (single input) or directory (multiple inputs)
```

Examples and notes:

```bash
# 1) Export JSON (default format=json). With multiple RDB inputs, the output is a JSON array:
$ ./rdr export -o out/report.json a.rdb b.rdb
# Output: exported JSON -> out/report.json

# 2) Export HTML (single input). If -o ends with .html, output that file; static assets go to sibling static/:
$ ./rdr export -f html -o out/report.html a.rdb
# Output: exported HTML -> out/report.html
# Files: out/report.html and out/static/...

# 3) Export HTML (multiple inputs). Prefer using a directory for -o; each RDB gets a .html; static assets go to static/ under that directory:
$ ./rdr export -f html -o out/reports a.rdb b.rdb
# Output:
# exported HTML -> out/reports/a.rdb.html
# exported HTML -> out/reports/b.rdb.html
# Files: out/reports/static/...

# 4) Notes:
# - --out/-o is required.
# - --format/-f supports json or html, default json.
# - When -o does not exist:
#   * If it ends with .html, treat as a file (create its parent directory if needed).
#   * Otherwise treat as a directory and create it.
# - If -o is an existing directory, or there are multiple inputs, export as a directory (one .html per input).
# - HTML export writes static assets to static/ for offline viewing.
```

### keys: print all keys

```text
NAME:
   rdr keys - get all keys from rdbfile

USAGE:
   rdr keys FILE1 [FILE2] [FILE3]...
```

Example:

```bash
$ ./rdr keys example.rdb
portfolio:stock_follower_count:ZH314136
portfolio:stock_follower_count:ZH654106
portfolio:stock_follower:ZH617824
portfolio:stock_follower_count:ZH001019
portfolio:stock_follower_count:ZH346349
portfolio:stock_follower_count:ZH951803
portfolio:stock_follower:ZH924804
portfolio:stock_follower_count:INS104806
```

## License

This project is under the Apache v2 License. See the [LICENSE](LICENSE) file for the full license text.
