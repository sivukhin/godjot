package djot_html

import (
	_ "embed"
	"testing"

	. "git.sr.ht/~ser/godjot/v2/djot_parser"
)

//go:embed bench/sample01.djot
var sample01 []byte

func BenchmarkConvertDjotToHtml(b *testing.B) {
	b.SetBytes(int64(len(sample01)))

	ast := BuildDjotAst(sample01)
	context := New()
	for i := 0; i < b.N; i++ {
		html := context.ConvertDjot(&HtmlWriter{}, ast...).String()
		if len(html) < 100 {
			b.Fail()
		}
	}
}
