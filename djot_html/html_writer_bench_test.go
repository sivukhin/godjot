package html_writer

import (
	_ "embed"
	"testing"

	. "github.com/sivukhin/godjot/djot_parser"
)

//go:embed bench/sample01.djot
var sample01 []byte

func BenchmarkConvertDjotToHtml(b *testing.B) {
	b.SetBytes(int64(len(sample01)))

	ast := BuildDjotAst(sample01)
	context := NewHtmlConversionContext("html")
	for i := 0; i < b.N; i++ {
		html := ConvertDjotToHtml(context, &HtmlWriter{}, ast...)
		if len(html) < 100 {
			b.Fail()
		}
	}
}
