package tokenizer

import (
	"bytes"
	"slices"
)

type TextReader struct {
	Text     []byte
	Position int
}

func (r TextReader) Peek() (byte, bool) {
	if r.Position < len(r.Text) {
		return r.Text[r.Position], true
	}
	return 0, false
}

func (r TextReader) MatchWhileSet(set string, next *TextReader) bool {
	bytesSet := []byte(set)
	advance := 0
	for r.Position+advance < len(r.Text) {
		current := r.Text[r.Position+advance]
		if !slices.Contains(bytesSet, current) {
			break
		}
		advance++
	}
	if advance == 0 {
		return false
	}
	*next = TextReader{Text: r.Text, Position: r.Position + advance}
	return true
}

func (r TextReader) MatchAny() TextReader {
	if r.Position == len(r.Text) {
		return r
	}
	return TextReader{Text: r.Text, Position: r.Position + 1}
}

func (r TextReader) MatchSingle(x byte, next *TextReader) bool {
	if current, ok := r.Peek(); ok && current == x {
		*next = r.MatchAny()
		return true
	}
	return false
}
func (r TextReader) Match(s string, next *TextReader) bool {
	ok := r.HasNext(s)
	if ok {
		*next = TextReader{Text: r.Text, Position: r.Position + len([]byte(s))}
	}
	return ok
}

func (r TextReader) Empty() bool           { return r.Position >= len(r.Text) }
func (r TextReader) HasPrev(s string) bool { return bytes.HasSuffix(r.Text[:r.Position], []byte(s)) }
func (r TextReader) HasNext(s string) bool { return bytes.HasPrefix(r.Text[r.Position:], []byte(s)) }
