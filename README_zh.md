# RDR: Redis 数据解析工具

RDR（redis data reveal）是一个用于解析 Redis RDB 文件的工具。相较于 [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools)，RDR 由 Go 语言实现，速度更快（在我的机器上，5GB RDB 文件约需 2 分钟）。

## 从源码构建（Windows / Linux）

需要 Go 1.21+（推荐使用 `go.mod` 中声明的 toolchain 1.23.2）。首次构建需先安装并执行资源打包工具。

```bash
# 1) 安装 go-bindata（保证 go-bindata 在 PATH 中）
$ go install github.com/go-bindata/go-bindata/...@latest

# 2) 拉取依赖
$ go mod tidy 

# 3) 生成内嵌静态资源与模板（必须执行）
$ go generate ./...
```

### Linux 本地构建

```bash
$ go build -o bin/rdr .
```

### Windows 本地构建（在 Windows 环境）

```powershell
PS> go build -o bin/rdr.exe .
```

### 交叉编译示例

```bash
# 从 Linux 构建 Windows 可执行文件
$ GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/rdr-windows-amd64.exe .

# 从 Linux 构建 Linux/amd64 与 Linux/arm64 可执行文件
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/rdr-linux-amd64 .
$ GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/rdr-linux-arm64 .

# 从 Windows 构建 Linux 可执行文件（PowerShell 示例）
PS> setx GOOS linux
PS> setx GOARCH amd64
PS> setx CGO_ENABLED 0
PS> go build -o bin/rdr-linux-amd64 .
```

提示：如果未执行 `go generate ./...`，将缺少 `static/static.go` 与 `views/views.go`，导致 `go build` 失败。

## 下载

最新版二进制会随 GitHub Releases 发布。以下直链会始终指向最新版本：

- Linux amd64: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-linux-amd64
- Linux arm64: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-linux-arm64
- macOS amd64: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-darwin-amd64
- macOS arm64（Apple Silicon）: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-darwin-arm64
- Windows amd64（.exe）: https://github.com/ericdong2012/rdr/releases/latest/download/rdr-windows-amd64.exe

校验文件： https://github.com/ericdong2012/rdr/releases/latest/download/checksums.sha256

## 使用方法

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

### show：Web 页面展示统计信息

```text
NAME:
   rdr show - show statistical information of rdbfile by webpage

USAGE:
   rdr show [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --port value, -p value  Port for rdr to listen (default: 8080)
```

示例：

```bash
$ ./rdr show -p 8080 *.rdb
```

注意：内存使用数据为近似值。

![show example](https://yqfile.alicdn.com/img_9bc93fc3a6b976fdf862c8314e34f454.png)

### export：导出统计信息到本地 JSON 或 HTML

```text
NAME:
   rdr export - export statistical information of rdbfile to local JSON or HTML

USAGE:
   rdr export [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --format value, -f value  export format: json or html (default: "json")
   --out value, -o value     output file path (single input) or directory (multiple inputs)
```

示例与说明：

```bash
# 1) 导出为 JSON（默认 format=json）。当有多个 RDB 输入时，输出为一个 JSON 数组：
$ ./rdr export -o out/report.json a.rdb b.rdb
# 输出：exported JSON -> out/report.json

# 2) 导出为 HTML（单输入）。当输出是具体文件名（以 .html 结尾）时，导出到该文件；静态资源会写到同目录下的 static/：
$ ./rdr export -f html -o out/report.html a.rdb
# 输出：exported HTML -> out/report.html
# 生成：out/report.html 以及 out/static/...

# 3) 导出为 HTML（多输入）。当有多个输入时，-o 建议指定目录；每个 RDB 会生成一个同名 .html，静态资源写到该目录下的 static/：
$ ./rdr export -f html -o out/reports a.rdb b.rdb
# 输出：
# exported HTML -> out/reports/a.rdb.html
# exported HTML -> out/reports/b.rdb.html
# 生成：out/reports/static/...

# 4) 特殊说明：
# - 必须提供 --out/-o 参数；未提供会报错。
# - --format/-f 支持 json 或 html，默认 json。
# - 当 -o 指向的路径不存在：
#   * 若以 .html 结尾，视为文件（将创建其父目录后写入该文件）。
#   * 其他情况视为目录并自动创建。
# - 当 -o 是已存在的目录，或输入文件数量>1 时，按目录导出（每个输入一个 .html）。
# - HTML 导出时会同时导出静态资源到目标目录下的 static/ 以便离线查看。
```

### keys：打印所有键

```text
NAME:
   rdr keys - get all keys from rdbfile

USAGE:
   rdr keys FILE1 [FILE2] [FILE3]...
```

示例：

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

## 许可证

本项目使用 Apache v2 许可证。详见 [LICENSE](LICENSE)。
