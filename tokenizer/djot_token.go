package tokenizer

import "fmt"

type DjotToken int

const (
	None          = 0
	DocumentBlock = 2*iota + 1
	HeadingBlock
	QuoteBlock
	ListItemBlock
	CodeBlock
	DivBlock
	PipeTableBlock
	ReferenceDefBlock
	FootnoteDefBlock
	ParagraphToken
	ThematicBreakToken

	Attribute

	Verbatim
	Escaped
	Span
	LinkUrl
	LinkReference
	Autolink
	Emphasis
	Strong
	Highlighted
	Subscript
	Superscript
	Insert
	Delete
	FootnoteReference
	Symbols
	RawFormat
)

func (t DjotToken) String() string {
	if t == None {
		return "None"
	}
	if t&1 == 0 {
		return (t ^ 1).String() + "Close"
	}
	switch t {
	case None:
		return "None"
	case DocumentBlock:
		return "DocumentBlock"
	case HeadingBlock:
		return "HeadingBlock"
	case QuoteBlock:
		return "QuoteBlock"
	case ListItemBlock:
		return "ListItemBlock"
	case CodeBlock:
		return "CodeBlock"
	case DivBlock:
		return "DivBlock"
	case PipeTableBlock:
		return "PipeTableBlock"
	case ReferenceDefBlock:
		return "ReferenceDefBlock"
	case FootnoteDefBlock:
		return "FootnoteDefBlock"
	case ParagraphToken:
		return "ParagraphToken"
	case ThematicBreakToken:
		return "ThematicBreakToken"
	case Attribute:
		return "Attribute"
	case Verbatim:
		return "Verbatim"
	case Escaped:
		return "Escaped"
	case Span:
		return "Span"
	case LinkUrl:
		return "LinkUrl"
	case LinkReference:
		return "LinkReference"
	case Autolink:
		return "Autolink"
	case Emphasis:
		return "Emphasis"
	case Strong:
		return "Strong"
	case Highlighted:
		return "Highlighted"
	case Subscript:
		return "Subscript"
	case Superscript:
		return "Superscript"
	case Insert:
		return "Insert"
	case Delete:
		return "Delete"
	case FootnoteReference:
		return "FootnoteReference"
	case Symbols:
		return "Symbols"
	case RawFormat:
		return "RawFormat"
	default:
		panic(fmt.Errorf("unexpected djot token type: %d", t))
	}
}
