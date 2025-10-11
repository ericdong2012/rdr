package dump

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "github.com/urfave/cli"
    "github.com/xueqiu/rdr/decoder"
    "github.com/xueqiu/rdr/static"
)

// Export parses rdb files and exports result to local JSON or HTML files.
func Export(c *cli.Context) {
    if c.NArg() < 1 {
        fmt.Fprintln(c.App.ErrWriter, "export requires at least 1 argument")
        cli.ShowCommandHelp(c, "export")
        return
    }

    format := strings.ToLower(c.String("format"))
    out := c.String("out")
    if format == "" {
        format = "json"
    }
    if out == "" {
        fmt.Fprintln(c.App.ErrWriter, "--out is required for export")
        return
    }

    switch format {
    case "json":
        exportJSON(c, out)
    case "html":
        exportHTML(c, out)
    default:
        fmt.Fprintf(c.App.ErrWriter, "unsupported format: %s (use json or html)\n", format)
    }
}

func exportJSON(cliCtx *cli.Context, outPath string) {
    nargs := cliCtx.NArg()
    results := make([]map[string]interface{}, 0, nargs)
    for i := 0; i < nargs; i++ {
        file := cliCtx.Args().Get(i)
        dec := decoder.NewDecoder()
        go Decode(cliCtx, dec, file)
        cnt := NewCounter()
        cnt.Count(dec.Entries)
        filename := filepath.Base(file)
        data := getData(filename, cnt)
        data["MemoryUse"] = dec.GetUsedMem()
        data["CTime"] = dec.GetTimestamp()
        results = append(results, data)
    }

    b, err := json.MarshalIndent(results, "", "    ")
    if err != nil {
        fmt.Fprintf(cliCtx.App.ErrWriter, "marshal json err: %v\n", err)
        return
    }
    if err := ioutil.WriteFile(outPath, b, 0644); err != nil {
        fmt.Fprintf(cliCtx.App.ErrWriter, "write json file err: %v\n", err)
        return
    }
    fmt.Fprintf(cliCtx.App.Writer, "exported JSON -> %s\n", outPath)
}

func exportHTML(cliCtx *cli.Context, out string) {
    // Ensure templates are initialized
    InitHTMLTmpl()

    // Prepare instances for template common data
    instances := []string{}

    // Determine output: if multiple inputs and out is a file, treat out as directory
    outStat, err := os.Stat(out)
    outIsDir := false
    if err == nil && outStat.IsDir() {
        outIsDir = true
    } else if err != nil {
        // if not exists and ends with .html, treat as file; otherwise treat as directory and create it
        if strings.HasSuffix(strings.ToLower(out), ".html") {
            // parent dir must exist
            if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
                fmt.Fprintf(cliCtx.App.ErrWriter, "create output dir err: %v\n", err)
                return
            }
        } else {
            if mkErr := os.MkdirAll(out, 0755); mkErr != nil {
                fmt.Fprintf(cliCtx.App.ErrWriter, "create output dir err: %v\n", mkErr)
                return
            }
            outIsDir = true
        }
    }

    // Export static assets alongside the HTML(s)
    var outDir string
    if outIsDir {
        outDir = out
    } else {
        outDir = filepath.Dir(out)
    }
    // put static assets under outDir/static
    staticTarget := filepath.Join(outDir, "static")
    if err := restoreStaticAssets(staticTarget, ""); err != nil {
        fmt.Fprintf(cliCtx.App.ErrWriter, "export static assets err: %v\n", err)
        return
    }

    nargs := cliCtx.NArg()
    for i := 0; i < nargs; i++ {
        file := cliCtx.Args().Get(i)
        filename := filepath.Base(file)
        instances = append(instances, filename)
    }

    for i := 0; i < nargs; i++ {
        file := cliCtx.Args().Get(i)
        dec := decoder.NewDecoder()
        go Decode(cliCtx, dec, file)
        cnt := NewCounter()
        cnt.Count(dec.Entries)

        // build data like rdbReveal
        data := map[string]interface{}{}
        // copy common data
        for k, v := range tplCommonData {
            data[k] = v
        }
        filename := filepath.Base(file)
        data["Instances"] = instances
        data["CurrentInstance"] = filename
        data["LargestKeys"] = cnt.GetLargestEntries(100)

        largestKeyPrefixesByType := map[string][]*PrefixEntry{}
        for _, entry := range cnt.GetLargestKeyPrefixes() {
            if entry.Bytes < 1000*1000 && len(largestKeyPrefixesByType[entry.Type]) > 50 {
                continue
            }
            largestKeyPrefixesByType[entry.Type] = append(largestKeyPrefixesByType[entry.Type], entry)
        }
        data["LargestKeyPrefixes"] = largestKeyPrefixesByType
        data["TypeBytes"] = cnt.typeBytes
        data["TypeNum"] = cnt.typeNum
        totleNum := uint64(0)
        for _, v := range cnt.typeNum {
            totleNum += v
        }
        totleBytes := uint64(0)
        for _, v := range cnt.typeBytes {
            totleBytes += v
        }
        data["TotleNum"] = totleNum
        data["TotleBytes"] = totleBytes
        lenLevelCount := map[string][]*PrefixEntry{}
        for _, entry := range cnt.GetLenLevelCount() {
            lenLevelCount[entry.Type] = append(lenLevelCount[entry.Type], entry)
        }
        data["LenLevelCount"] = lenLevelCount

        // render
        htmlBytes, err := RenderHTML("base.html", "revel.html", data)
        if err != nil {
            fmt.Fprintf(cliCtx.App.ErrWriter, "render html err: %v\n", err)
            return
        }
        // adjust /static/ to relative static/
        htmlBytes = bytes.ReplaceAll(htmlBytes, []byte("=\"/static/"), []byte("=\"static/"))
        htmlBytes = bytes.Replace(htmlBytes, []byte("='/static/"), []byte("='static/"), -1)

        // determine output file
        var outFile string
        if outIsDir {
            outFile = filepath.Join(out, filename+".html")
        } else {
            // single input -> out file as given
            outFile = out
        }
        if err := ioutil.WriteFile(outFile, htmlBytes, 0644); err != nil {
            fmt.Fprintf(cliCtx.App.ErrWriter, "write html file err: %v\n", err)
            return
        }
        fmt.Fprintf(cliCtx.App.Writer, "exported HTML -> %s\n", outFile)

        // if out is a single file but there are multiple inputs, we only write first
        if !outIsDir {
            break
        }
    }
}

// restoreStaticAssets writes embedded static assets to targetDir preserving paths.
func restoreStaticAssets(targetDir, name string) error {
    children, err := static.AssetDir(name)
    if err != nil {
        // file
        data, err := static.Asset(name)
        if err != nil {
            return err
        }
        info, err := static.AssetInfo(name)
        if err != nil {
            return err
        }
        if err := os.MkdirAll(filepath.Join(targetDir, filepath.Dir(name)), 0755); err != nil {
            return err
        }
        if err := ioutil.WriteFile(filepath.Join(targetDir, name), data, info.Mode()); err != nil {
            return err
        }
        return nil
    }
    for _, child := range children {
        if err := restoreStaticAssets(targetDir, filepath.Join(name, child)); err != nil {
            return err
        }
    }
    return nil
}
