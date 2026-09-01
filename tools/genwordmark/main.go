//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukesampson/figlet/figletlib"
)

func main() {
	fontDir := filepath.Join("internal", "ds", "fonts")
	if err := os.MkdirAll(fontDir, 0o755); err != nil {
		panic(err)
	}

	for _, font := range []string{"Big", "Standard", "Slant"} {
		url := "https://raw.githubusercontent.com/xero/figlet-fonts/master/" + font + ".flf"
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		fontPath := filepath.Join(fontDir, strings.ToLower(font)+".flf")
		f, err := os.Create(fontPath)
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			panic(err)
		}
		resp.Body.Close()
		if err := f.Close(); err != nil {
			panic(err)
		}

		ff, err := figletlib.GetFontByName(fontDir, strings.ToLower(font))
		if err != nil {
			panic(err)
		}
		var b strings.Builder
		figletlib.FPrintMsg(&b, "termcord", ff, 200, ff.Settings(), "left")
		out := strings.TrimRight(b.String(), "\n")
		if strings.Contains(out, "888") {
			fmt.Printf("=== %s (skip: has 888) ===\n", font)
			continue
		}
		fmt.Printf("=== %s ===\n%s\n\n", font, out)
	}
}
