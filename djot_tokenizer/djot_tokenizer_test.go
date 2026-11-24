package djot_tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"git.sr.ht/~ser/godjot/v2/tokenizer"
)

func TestSimpleText(t *testing.T) {
	tokens := BuildDjotTokens([]byte("hello *world*!"))
	t.Log(tokens)
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 8},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 6},
		{Type: None, Start: 0, End: 6},
		{Type: StrongInline, Start: 6, End: 7, JumpToPair: 2},
		{Type: None, Start: 7, End: 12},
		{Type: StrongInline ^ tokenizer.Open, Start: 12, End: 13, JumpToPair: -2},
		{Type: None, Start: 13, End: 14},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 14, End: 14, JumpToPair: -6},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 14, End: 14, JumpToPair: -8},
	}, tokens)
}

func TestSimpleDocument(t *testing.T) {
	bytes := []byte(`## This is *multiline*
header

Then we have simple paragraph
> This is not quote!

> But this is quote with item list
> 
> 1. one-line item
> 2. multi-line
> item
> 3. last item`)
	tokens := BuildDjotTokens(bytes)
	for _, token := range tokens {
		t.Log(token)
	}
}

func TestSimpleLink(t *testing.T) {
	tokens := BuildDjotTokens([]byte("[a](b)"))
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 9},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 7},
		{Type: SpanInline, Start: 0, End: 1, JumpToPair: 2},
		{Type: None, Start: 1, End: 2},
		{Type: SpanInline ^ tokenizer.Open, Start: 2, End: 3, JumpToPair: -2},
		{Type: LinkUrlInline, Start: 3, End: 4, JumpToPair: 2},
		{Type: None, Start: 4, End: 5},
		{Type: LinkUrlInline ^ tokenizer.Open, Start: 5, End: 6, JumpToPair: -2},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 6, End: 6, JumpToPair: -7},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 6, End: 6, JumpToPair: -9},
	}, tokens)
}

func TestSimpleLinkWithNewline(t *testing.T) {
	tokens := BuildDjotTokens([]byte("[My link text](http://example.com?product_number=234234234234\n234234234234)"))
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 11},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 9},
		{Type: SpanInline, Start: 0, End: 1, JumpToPair: 2},
		{Type: None, Start: 1, End: 13},
		{Type: SpanInline ^ tokenizer.Open, Start: 13, End: 14, JumpToPair: -2},
		{Type: LinkUrlInline, Start: 14, End: 15, JumpToPair: 4},
		{Type: None, Start: 15, End: 61},
		{Type: SmartSymbolInline, Start: 61, End: 62},
		{Type: None, Start: 62, End: 74},
		{Type: LinkUrlInline ^ tokenizer.Open, Start: 74, End: 75, JumpToPair: -4},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 75, End: 75, JumpToPair: -9},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 75, End: 75, JumpToPair: -11},
	}, tokens)
}

func TestMathVerbatim(t *testing.T) {
	tokens := BuildDjotTokens([]byte("$$`1+1=2`"))
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 6},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 4},
		{Type: VerbatimInline, Start: 0, End: 3, JumpToPair: 2, Attributes: tokenizer.NewAttributes(tokenizer.AttributeEntry{
			Key: DisplayMathKey,
		})},
		{Type: None, Start: 3, End: 8},
		{Type: VerbatimInline ^ tokenizer.Open, Start: 8, End: 9, JumpToPair: -2},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 9, End: 9, JumpToPair: -4},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 9, End: 9, JumpToPair: -6},
	}, tokens)
}

func TestVerbatim(t *testing.T) {
	tokens := BuildDjotTokens([]byte("``VerbatimInline with a backtick` character``\n`VerbatimInline with three backticks ``` character`"))
	t.Log(tokens)
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 10},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 8},
		{Type: VerbatimInline, Start: 0, End: 2, JumpToPair: 2},
		{Type: None, Start: 2, End: 43},
		{Type: VerbatimInline ^ tokenizer.Open, Start: 43, End: 45, JumpToPair: -2},
		{Type: SmartSymbolInline, Start: 45, End: 46},
		{Type: VerbatimInline, Start: 46, End: 47, JumpToPair: 2},
		{Type: None, Start: 47, End: 96},
		{Type: VerbatimInline ^ tokenizer.Open, Start: 96, End: 97, JumpToPair: -2},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 97, End: 97, JumpToPair: -8},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 97, End: 97, JumpToPair: -10},
	}, tokens)
}

func TestNestedEmphasis(t *testing.T) {
	tokens := BuildDjotTokens([]byte("___abc___"))
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 10},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 8},
		{Type: EmphasisInline, Start: 0, End: 1, JumpToPair: 6},
		{Type: EmphasisInline, Start: 1, End: 2, JumpToPair: 4},
		{Type: EmphasisInline, Start: 2, End: 3, JumpToPair: 2},
		{Type: None, Start: 3, End: 6},
		{Type: EmphasisInline ^ tokenizer.Open, Start: 6, End: 7, JumpToPair: -2},
		{Type: EmphasisInline ^ tokenizer.Open, Start: 7, End: 8, JumpToPair: -4},
		{Type: EmphasisInline ^ tokenizer.Open, Start: 8, End: 9, JumpToPair: -6},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 9, End: 9, JumpToPair: -8},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 9, End: 9, JumpToPair: -10},
	}, tokens)
}

func TestUnmatchedEmphasis(t *testing.T) {
	tokens := BuildDjotTokens([]byte("___ (not an emphasized `_` character)"))
	require.Equal(t, tokenizer.TokenList[DjotToken]{
		{Type: DocumentBlock, Start: 0, End: 0, JumpToPair: 8},
		{Type: ParagraphBlock, Start: 0, End: 0, JumpToPair: 6},
		{Type: None, Start: 0, End: 23},
		{Type: VerbatimInline, Start: 23, End: 24, JumpToPair: 2},
		{Type: None, Start: 24, End: 25},
		{Type: VerbatimInline ^ tokenizer.Open, Start: 25, End: 26, JumpToPair: -2},
		{Type: None, Start: 26, End: 37},
		{Type: ParagraphBlock ^ tokenizer.Open, Start: 37, End: 37, JumpToPair: -6},
		{Type: DocumentBlock ^ tokenizer.Open, Start: 37, End: 37, JumpToPair: -8},
	}, tokens)
}
