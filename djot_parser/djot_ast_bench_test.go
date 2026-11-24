package djot_parser

import (
	"bytes"
	_ "embed"
	"math/rand"
	"os"
	"testing"

	"git.sr.ht/~ser/godjot/v2/djot_tokenizer"
)

//go:embed bench/sample01.djot
var sample01 []byte

//go:embed bench/sample02.djot
var sample02 []byte

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
	_ = os.WriteFile("bench/sample02-new.djot", anonymize(sample02), 0660)
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
	b.Run("sample01", func(b *testing.B) {
		b.SetBytes(int64(len(sample01)))
		for i := 0; i < b.N; i++ {
			ast := BuildDjotAst(sample01)
			if len(ast) == 0 {
				b.Fail()
			}
		}
	})
	b.Run("sample02", func(b *testing.B) {
		b.SetBytes(int64(len(sample02)))
		for i := 0; i < b.N; i++ {
			ast := BuildDjotAst(sample02)
			if len(ast) == 0 {
				b.Fail()
			}
		}
	})
}
