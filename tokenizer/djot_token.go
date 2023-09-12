package tokenizer

import (
	"fmt"
	"unicode"
)

type DjotToken int

const (
	Close            = 1
	firstSimpleToken = Emphasis
	lastSimpleToken  = Attribute
)

const (
	Raw DjotToken = iota * 2
	Root
	Doc
	Emphasis
	Strong
	Highlighted
	Subscript
	Superscript
	Insert
	Delete
	Autolink
	Span
	RawFormat
	Symbols
	Verbatim
	Attribute

	String
	Escaped
)

func (t DjotToken) String() string {
	if t&1 != 0 {
		return (t ^ 1).String() + "Close"
	}
	switch t {
	case Raw:
		return "Raw"
	case Root:
		return "Root"
	case Doc:
		return "Doc"
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
	case Autolink:
		return "Autolink"
	case Span:
		return "Span"
	case Attribute:
		return "Attribute"
	case RawFormat:
		return "RawFormat"
	case Symbols:
		return "Symbols"
	case Verbatim:
		return "Verbatim"
	case String:
		return "String"
	case Escaped:
		return "Escaped"
	default:
		panic(fmt.Errorf("unexpected token type: %d", t))
	}
}

func MatchDjotToken(r TextReader, token DjotToken, next *TextReader) bool {
	switch token {
	case Emphasis:
		return r.Match("_", next) && !next.HasNext(" ") || r.Match("{_", next)
	case Emphasis | Close:
		return r.Match("_", next) && !r.HasPrev(" ") || r.Match("_}", next)
	case Strong:
		return r.Match("*", next) && !next.HasNext(" ") || r.Match("{*", next)
	case Strong | Close:
		return r.Match("*", next) && !r.HasPrev(" ") || r.Match("*}", next)
	case Highlighted:
		return r.Match("{=", next)
	case Highlighted | Close:
		return r.Match("=}", next)
	case Superscript:
		return r.Match("^", next) || r.Match("{^", next)
	case Superscript | Close:
		return r.Match("^", next) || r.Match("^}", next)
	case Subscript:
		return r.Match("~", next) || r.Match("{~", next)
	case Subscript | Close:
		return r.Match("~", next) || r.Match("~}", next)
	case Insert:
		return r.Match("{+", next)
	case Insert | Close:
		return r.Match("+}", next)
	case Delete:
		return r.Match("{-", next)
	case Delete | Close:
		return r.Match("-}", next)
	case Span:
		return r.Match("[^", next) || r.Match("![", next) || r.Match("[", next)
	case Span | Close:
		return r.Match("]", next)
	case Autolink:
		return r.Match("<", next)
	case Autolink | Close:
		return r.Match(">", next)
	case Attribute:
		return r.Match("{", next)
	case RawFormat:
		return r.HasPrev("`") && r.Match("{=", next)
	case RawFormat | Close, Attribute | Close:
		return r.Match("}", next)
	case Symbols, Symbols | Close:
		return r.Match(":", next)
	case Verbatim, Verbatim | Close:
		hasPrefix := r.Match("$$", next) || r.Match("$", next)
		return (hasPrefix && next.MatchWhileSet("`", next)) || r.MatchWhileSet("`", next)
	case String, String | Close:
		return r.Match("\"", next)
	case Escaped:
		if !r.Match("\\", next) {
			return false
		}
		if symbol, ok := next.Peek(); ok && unicode.IsPunct(rune(symbol)) {
			*next = next.MatchAny()
			return true
		}
		return next.MatchWhileSet(" \t", next) && next.Match("\n", next)
	default:
		panic(fmt.Errorf("unexpected djot token: %v", token))
	}
}
