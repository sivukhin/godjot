package djot_html

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/sivukhin/godjot/djot_parser"
	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/tokenizer"
)

func printDjot(text string) string {
	document := []byte(text)
	ast := BuildDjotAst(document)
	fmt.Printf("ast: %v\n", ast)
	return New().ConvertDjot(&HtmlWriter{}, ast...).String()
}

const examplesDir = "examples"

func TestDownloadExample(t *testing.T) {
	normalize := func(line string) string {
		line = strings.Trim(line, "\r\n\t")
		line = strings.TrimPrefix(line, "<pre><code>")
		line = strings.TrimSuffix(line, "</code></pre>")
		return line
	}

	response, err := http.Get("https://raw.githubusercontent.com/jgm/djot/main/doc/syntax.html")
	require.Nil(t, err)
	docBytes, err := io.ReadAll(response.Body)
	require.Nil(t, err)
	var (
		djotStartToken = []byte(`<div class="djot">`)
		htmlStartToken = []byte(`<div class="html">`)
		endToken       = []byte(`</div>`)
	)
	example := 0
	for {
		djotStart := bytes.Index(docBytes, djotStartToken)
		if djotStart == -1 {
			break
		}
		djotEnd := djotStart + bytes.Index(docBytes[djotStart:], endToken)
		djotExample := html.UnescapeString(normalize(string(docBytes[djotStart+len(djotStartToken) : djotEnd])))
		docBytes = docBytes[djotEnd+len(endToken):]

		htmlStart := bytes.Index(docBytes, htmlStartToken)
		require.NotEqual(t, htmlStart, -1)
		htmlEnd := htmlStart + bytes.Index(docBytes[htmlStart:], endToken)
		htmlExample := html.UnescapeString(normalize(string(docBytes[htmlStart+len(htmlStartToken) : htmlEnd])))
		docBytes = docBytes[htmlEnd+len(endToken):]

		// Ignore 64th example because it's not self-contained and requires additional definition of table
		if example != 64 {
			require.Nil(t, os.WriteFile(path.Join(examplesDir, fmt.Sprintf("%02d.html", example)), []byte(htmlExample), 0660))
			require.Nil(t, os.WriteFile(path.Join(examplesDir, fmt.Sprintf("%02d.djot", example)), []byte(djotExample), 0660))
		}
		example++
	}
}

func TestStartSymbol(t *testing.T) {
	dir, err := os.ReadDir(examplesDir)
	require.Nil(t, err)
	for _, entry := range dir {
		name := entry.Name()
		example, ok := strings.CutSuffix(name, ".html")
		if !ok {
			continue
		}
		djotExample, err := os.ReadFile(path.Join(examplesDir, fmt.Sprintf("%v.djot", example)))
		require.Nil(t, err)
		_ = BuildDjotAst(djotExample)
	}
	symbols := make([]byte, 0)
	for s := range djot_tokenizer.StartSymbols {
		if !tokenizer.SpaceNewLineByteMask.Has(s) {
			symbols = append(symbols, s)
		}
	}
	sort.Slice(symbols, func(i, j int) bool { return symbols[i] < symbols[j] })
	t.Logf("%#v", string(symbols))
}

func TestDjotDocExample(t *testing.T) {
	dir, err := os.ReadDir(examplesDir)
	require.Nil(t, err)
	for _, entry := range dir {
		name := entry.Name()
		example, ok := strings.CutSuffix(name, ".html")
		if !ok {
			continue
		}
		htmlExample, err := os.ReadFile(path.Join(examplesDir, fmt.Sprintf("%v.html", example)))
		require.Nil(t, err)
		djotExample, err := os.ReadFile(path.Join(examplesDir, fmt.Sprintf("%v.djot", example)))
		require.Nil(t, err)
		t.Run(example+":"+string(djotExample), func(t *testing.T) {
			result := printDjot(string(djotExample))
			require.Equalf(
				t, string(htmlExample), result,
				"invalid html (%v != %v), djot tokens: %v",
				string(htmlExample), result,
				djot_tokenizer.BuildDjotTokens(djotExample),
			)
		})
	}
}

func TestManualExamples(t *testing.T) {
	t.Run("link in text", func(t *testing.T) {
		result := printDjot("link http://localhost:3000/debug/pprof/profile?seconds=10 -o profile.pprof")
		require.Equal(t, "<p>link http://localhost:3000/debug/pprof/profile?seconds=10 -o profile.pprof</p>\n", result)
	})
	t.Run("block attributes", func(t *testing.T) {
		result := printDjot(`{key="value"}
# Header`)
		require.Equal(t, `<section id="Header">
<h1 key="value">Header</h1>
</section>
`, result)
	})
	t.Run("inline attributes", func(t *testing.T) {
		result := printDjot(`![img](link){key="value"}`)
		t.Log(result)
	})
}
