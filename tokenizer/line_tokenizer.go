package tokenizer

import (
	"bytes"
)

type LineTokenizer struct {
	Document  []byte
	docOffset int
}

var newline = []byte("\n")

func (tokenizer *LineTokenizer) Scan() (start, end int, eof bool) {
	if tokenizer.docOffset == len(tokenizer.Document) {
		eof = true
		return
	}

	suffix := tokenizer.Document[tokenizer.docOffset:]
	newlineIndex := bytes.Index(suffix, newline)
	if newlineIndex == -1 {
		start = tokenizer.docOffset
		end = len(tokenizer.Document)
		tokenizer.docOffset = len(tokenizer.Document)
		return
	}
	start = tokenizer.docOffset
	end = tokenizer.docOffset + newlineIndex + 1
	tokenizer.docOffset += newlineIndex + 1
	return
}
