// Copyright 2017 XUEQIU.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/urfave/cli"

	"fmt"
	"time"

	"github.com/xueqiu/rdr/decoder"
	"github.com/xueqiu/rdr/dump"
)

//go:generate go-bindata -prefix "static/" -o=static/static.go -pkg=static -ignore static.go static/...
//go:generate go-bindata -prefix "views/" -o=views/views.go -pkg=views -ignore views.go views/...

// keys is function for command `keys`
// output all keys in rdbfile(s) get from args
func keys(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Fprintln(c.App.ErrWriter, "keys requires at least 1 argument")
		cli.ShowCommandHelp(c, "keys")
		return
	}
	withExpire := c.Bool("with-expire")
	for _, filepath := range c.Args() {
		decoder := decoder.NewDecoder()
		go dump.Decode(c, decoder, filepath)
		for e := range decoder.Entries {
			if withExpire {
				if e.ExpireAt > 0 {
					ts := time.Unix(e.ExpireAt/1000, (e.ExpireAt%1000)*int64(time.Millisecond))
					formatted := ts.Format("2006-01-02T15:04:05.000000")
					fmt.Fprintf(c.App.Writer, "%v, %s, %d, %s\n", e.Key, e.Type, e.Bytes, formatted)
				} else {
					fmt.Fprintf(c.App.Writer, "%v, %s, %d,\n", e.Key, e.Type, e.Bytes)
				}
			} else {
				fmt.Fprintf(c.App.Writer, "%v\n", e.Key)
			}
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "rdr"
	app.Usage = "a tool to parse redis rdbfile"
	app.Version = "v0.0.1"
	app.Writer = os.Stdout
	app.ErrWriter = os.Stderr
	app.Commands = []cli.Command{
		cli.Command{
			Name:      "dump",
			Usage:     "dump statistical information of rdbfile to STDOUT",
			ArgsUsage: "FILE1 [FILE2] [FILE3]...",
			Action:    dump.ToCliWriter,
		},
		cli.Command{
			Name:      "show",
			Usage:     "show statistical information of rdbfile by webpage",
			ArgsUsage: "DIR1 [DIR2] [DIR3] or FILE1 [FILE2] [FILE3]...",
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "port, p",
					Value: 8080,
					Usage: "Port for rdr to listen",
				},
			},
			Action: dump.Show,
		},
		cli.Command{
			Name:      "export",
			Usage:     "export statistical information of rdbfile to local JSON or HTML",
			ArgsUsage: "FILE1 [FILE2] [FILE3]...",
			UsageText: "rdr export [--format json|html] --out <file|dir> FILE1 [FILE2] [FILE3]...\n\nExample:\n  ./rdr export -f html -o out/report.html a.rdb b.rdb",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "format, f",
					Value: "json",
					Usage: "export format: json or html",
				},
				cli.StringFlag{
					Name:  "out, o",
					Usage: "output file path (for single input) or directory (for multiple inputs)",
				},
			},
			Action: dump.Export,
		},
		cli.Command{
			Name:      "keys",
			Usage:     "get all keys from rdbfile",
			ArgsUsage: "FILE1 [FILE2] [FILE3]...",
			UsageText: "rdr keys [--with-expire] FILE1 [FILE2] [FILE3]...\n\nExample:\n  ./rdr keys -e a.rdb b.rdb",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "with-expire, e",
					Usage: "print 'key, <type>, <size_in_bytes>, <expiry(2006-01-02T15:04:05.000000)>'; if no expiry prints trailing comma after size",
				},
			},
			Action:    keys,
		},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.ErrWriter, "command %q can not be found.\n", command)
		cli.ShowAppHelp(c)
	}
	app.Run(os.Args)
}
