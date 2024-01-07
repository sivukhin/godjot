package djot_parser

import (
	"bytes"
	_ "embed"
	"math/rand"
	"os"
	"testing"

	"github.com/sivukhin/godjot/djot_tokenizer"
)

//go:embed bench/sample01.djot
var sample01 []byte

func anonymize(data []byte) []byte {
	symbols := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := bytes.Clone(data)
	for i, b := range data {
		index := bytes.IndexByte(symbols, b)
		if index >= 0 {
			result[i] = symbols[rand.Intn(len(symbols))]
		}
	}
	return result
}

func TestAnonymize(t *testing.T) {
	t.Skip()
	_ = os.WriteFile("bench/sample01-new.djot", anonymize(sample01), 0660)
}

func BenchmarkBuildDjotTokens(b *testing.B) {
	b.SetBytes(int64(len(sample01)))
	for i := 0; i < b.N; i++ {
		tokens := djot_tokenizer.BuildDjotTokens(sample01)
		if len(tokens) == 0 {
			b.Fail()
		}
	}
}

func BenchmarkBuildDjotAst(b *testing.B) {
	b.SetBytes(int64(len(sample01)))
	for i := 0; i < b.N; i++ {
		ast := BuildDjotAst(sample01)
		if len(ast) == 0 {
			b.Fail()
		}
	}
}

func BenchmarkConvertDjotToHtml(b *testing.B) {
	b.SetBytes(int64(len(sample01)))

	ast := BuildDjotAst(sample01)
	context := NewConversionContext("html", DefaultConversionRegistry)
	for i := 0; i < b.N; i++ {
		html := context.ConvertDjotToHtml(ast...)
		if len(html) < 100 {
			b.Fail()
		}
	}
}
