package tokenizer

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleText(t *testing.T) {
	tokens := DjotTokens([]byte("hello *world*!"))
	t.Log(tokens)
	require.Equal(t, []Token[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 7},
		{Type: ParagraphToken, Start: 0, End: 0, JumpToPair: 5},
		{Type: None, Start: 0, End: 6},
		{Type: Strong, Start: 6, End: 7, JumpToPair: 2},
		{Type: None, Start: 7, End: 12},
		{Type: Strong ^ Open, Start: 12, End: 13, JumpToPair: -2},
		//{Type: None, Start: 13, End: 14},
		{Type: ParagraphToken ^ Open, Start: 14, End: 14, JumpToPair: -5},
		{Type: DocumentBlock ^ Open, Start: 14, End: 14, JumpToPair: -7},
	}, tokens)
}

func TestSimpleQuote(t *testing.T) {
	tokens := DjotTokens([]byte(`hello

2. world`))
	t.Log(tokens)
}

//
//func TestSimpleLink(t *testing.T) {
//	tokens := DjotTokens([]byte("[a](b)"))
//	require.Equal(t, []Token[DjotToken]{
//		{Type: Paragraph, Start: 0, End: 0, JumpToPair: 7},
//		{Type: Span, Start: 0, End: 1, JumpToPair: 2},
//		{Type: None, Start: 1, End: 2},
//		{Type: Span ^ Open, Start: 2, End: 3, JumpToPair: -2},
//		{Type: LinkUrl, Start: 3, End: 4, JumpToPair: 2},
//		{Type: None, Start: 4, End: 5},
//		{Type: LinkUrl ^ Open, Start: 5, End: 6, JumpToPair: -2},
//		{Type: Paragraph ^ Open, Start: 6, End: 6, JumpToPair: -7},
//	}, tokens)
//}
//
//func TestSimpleLinkWithNewline(t *testing.T) {
//	tokens := DjotTokens([]byte("[My link text](http://example.com?product_number=234234234234\n234234234234)"))
//	require.Equal(t, []Token[DjotToken]{
//		{Type: Paragraph, Start: 0, End: 0, JumpToPair: 7},
//		{Type: Span, Start: 0, End: 1, JumpToPair: 2},
//		{Type: None, Start: 1, End: 13},
//		{Type: Span ^ Open, Start: 13, End: 14, JumpToPair: -2},
//		{Type: LinkUrl, Start: 14, End: 15, JumpToPair: 2},
//		{Type: None, Start: 15, End: 74},
//		{Type: LinkUrl ^ Open, Start: 74, End: 75, JumpToPair: -2},
//		{Type: Paragraph ^ Open, Start: 75, End: 75, JumpToPair: -7},
//	}, tokens)
//}
//
//func TestMathVerbatim(t *testing.T) {
//	tokens := DjotTokens([]byte("$$`1+1=2`"))
//	require.Equal(t, []Token[DjotToken]{
//		{Type: Paragraph, Start: 0, End: 0, JumpToPair: 4},
//		{Type: Verbatim, Start: 0, End: 3, JumpToPair: 2},
//		{Type: None, Start: 3, End: 8},
//		{Type: Verbatim ^ Open, Start: 8, End: 9, JumpToPair: -2},
//		{Type: Paragraph ^ Open, Start: 9, End: 9, JumpToPair: -4},
//	}, tokens)
//}
//
//func TestVerbatim(t *testing.T) {
//	tokens := DjotTokens([]byte("``Verbatim with a backtick` character``\n`Verbatim with three backticks ``` character`"))
//	require.Equal(t, []Token[DjotToken]{
//		{Type: Paragraph, Start: 0, End: 0, JumpToPair: 8},
//		{Type: Verbatim, Start: 0, End: 2, JumpToPair: 2},
//		{Type: None, Start: 2, End: 37},
//		{Type: Verbatim ^ Open, Start: 37, End: 39, JumpToPair: -2},
//		{Type: None, Start: 39, End: 40},
//		{Type: Verbatim, Start: 40, End: 41, JumpToPair: 2},
//		{Type: None, Start: 41, End: 84},
//		{Type: Verbatim ^ Open, Start: 84, End: 85, JumpToPair: -2},
//		{Type: Paragraph ^ Open, Start: 85, End: 85, JumpToPair: -8},
//	}, tokens)
//}
//
//func TestNestedEmphasis(t *testing.T) {
//	tokens := DjotTokens([]byte("___abc___"))
//	require.Equal(t, []Token[DjotToken]{
//		{Type: Paragraph, Start: 0, End: 0, JumpToPair: 8},
//		{Type: Emphasis, Start: 0, End: 1, JumpToPair: 6},
//		{Type: Emphasis, Start: 1, End: 2, JumpToPair: 4},
//		{Type: Emphasis, Start: 2, End: 3, JumpToPair: 2},
//		{Type: None, Start: 3, End: 6},
//		{Type: Emphasis ^ Open, Start: 6, End: 7, JumpToPair: -2},
//		{Type: Emphasis ^ Open, Start: 7, End: 8, JumpToPair: -4},
//		{Type: Emphasis ^ Open, Start: 8, End: 9, JumpToPair: -6},
//		{Type: Paragraph ^ Open, Start: 9, End: 9, JumpToPair: -8},
//	}, tokens)
//}
//
//func TestUnmatchedEmphasis(t *testing.T) {
//	tokens := DjotTokens([]byte("___ (not an emphasized `_` character)"))
//	require.Equal(t, []Token[DjotToken]{
//		{Type: Paragraph, Start: 0, End: 0, JumpToPair: 6},
//		{Type: None, Start: 0, End: 23},
//		{Type: Verbatim, Start: 23, End: 24, JumpToPair: 2},
//		{Type: None, Start: 24, End: 25},
//		{Type: Verbatim ^ Open, Start: 25, End: 26, JumpToPair: -2},
//		{Type: None, Start: 26, End: 37},
//		{Type: Paragraph ^ Open, Start: 37, End: 37, JumpToPair: -6},
//	}, tokens)
//}
