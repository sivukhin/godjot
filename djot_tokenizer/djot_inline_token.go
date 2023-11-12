package djot_tokenizer

import (
	"fmt"
	"github.com/sivukhin/godjot/tokenizer"
	"unicode"
)

var (
	DollarByteMask             = tokenizer.NewByteMask([]byte("$"))
	BacktickByteMask           = tokenizer.NewByteMask([]byte("`"))
	SmartSymbolByteMask        = tokenizer.NewByteMask([]byte("\n'\""))
	AlphaNumericSymbolByteMask = tokenizer.NewByteMask([]byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"))
)

func MatchInlineToken(
	r tokenizer.TextReader,
	s tokenizer.ReaderState,
	tokenType DjotToken,
) (next tokenizer.ReaderState) {
	next = tokenizer.Unmatched

	switch tokenType {
	case ImageSpanInline:
		next = r.Token(s, "![")
	case SpanInline:
		next = r.Token(s, "[")
	case SpanInline ^ tokenizer.Open, ImageSpanInline ^ tokenizer.Open:
		next = r.Token(s, "]")
	case LinkUrlInline:
		next = r.Token(s, "(")
	case LinkUrlInline ^ tokenizer.Open:
		next = r.Token(s, ")")
	case LinkReferenceInline:
		next = r.Token(s, "[")
	case LinkReferenceInline ^ tokenizer.Open:
		next = r.Token(s, "]")
	case AutolinkInline:
		next = r.Token(s, "<")
	case AutolinkInline ^ tokenizer.Open:
		next = r.Token(s, ">")
	case VerbatimInline:
		next = r.MaskRepeat(s, DollarByteMask, 0)
		if next-s > 2 {
			break
		}
		next = r.MaskRepeat(next, BacktickByteMask, 1)
	case VerbatimInline ^ tokenizer.Open:
		next = r.MaskRepeat(s, BacktickByteMask, 1)
	case EmphasisInline:
		next = r.Token(s, "{_")
		if next.Matched() {
			break
		}
		next = r.Token(s, "_")
		if r.HasMask(next, tokenizer.SpaceNewLineByteMask) {
			next = tokenizer.Unmatched
		}
	case EmphasisInline ^ tokenizer.Open:
		next = r.Token(s, "_}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "_")
		if r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			next = tokenizer.Unmatched
		}
	case StrongInline:
		next = r.Token(s, "{*")
		if next.Matched() {
			break
		}
		next = r.Token(s, "*")
		if r.HasMask(next, tokenizer.SpaceNewLineByteMask) {
			next = tokenizer.Unmatched
		}
	case StrongInline ^ tokenizer.Open:
		next = r.Token(s, "*}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "*")
		if r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			next = tokenizer.Unmatched
		}
	case HighlightedInline:
		next = r.Token(s, "{=")
	case HighlightedInline ^ tokenizer.Open:
		next = r.Token(s, "=}")
	case SuperscriptInline:
		next = r.Token(s, "{^")
		if next.Matched() {
			break
		}
		next = r.Token(s, "^")
	case SuperscriptInline ^ tokenizer.Open:
		next = r.Token(s, "^}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "^")
	case SubscriptInline:
		next = r.Token(s, "{~")
		if next.Matched() {
			break
		}
		next = r.Token(s, "~")
	case SubscriptInline ^ tokenizer.Open:
		next = r.Token(s, "~}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "~")
	case InsertInline:
		next = r.Token(s, "{+")
	case InsertInline ^ tokenizer.Open:
		next = r.Token(s, "+}")
	case DeleteInline:
		next = r.Token(s, "{-")
	case DeleteInline ^ tokenizer.Open:
		next = r.Token(s, "-}")
	case FootnoteReferenceInline:
		next = r.Token(s, "[^")
	case FootnoteReferenceInline ^ tokenizer.Open:
		next = r.Token(s, "]")
	case EscapedSymbolInline:
		next = r.Token(s, "\\")
		if !next.Matched() {
			break
		}
		if r.Empty(next) {
			next = tokenizer.Unmatched
			break
		}
		if symbol := r[next]; unicode.IsPunct(rune(symbol)) {
			next++
		} else {
			next = r.MaskRepeat(next, tokenizer.SpaceByteMask, 0)
			next = r.Token(next, "\n")
		}
	case RawFormatInline:
		next = r.Token(s, "{=")
	case RawFormatInline ^ tokenizer.Open:
		next = r.Token(s, "}")
	case SymbolsInline:
		next = r.Token(s, ":")
		if !next.Matched() {
			break
		}
		if word := r.MaskRepeat(next, AlphaNumericSymbolByteMask, 0); !word.Matched() || !r.Token(word, ":").Matched() {
			next = tokenizer.Unmatched
		}
	case SymbolsInline ^ tokenizer.Open:
		next = r.Token(s, ":")
	case SmartSymbolInline:
		if !next.Matched() {
			next = r.Token(s, "{")
			if next.Matched() {
				next = r.Mask(next, SmartSymbolByteMask)
			}
		}
		if !next.Matched() {
			next = r.Mask(s, SmartSymbolByteMask)
			if next.Matched() {
				next = r.Token(next, "}")
			}
		}
		if !next.Matched() {
			next = r.Mask(s, SmartSymbolByteMask)
		}
		if !next.Matched() {
			next = r.Token(s, "...")
		}
		if !next.Matched() {
			next = r.ByteRepeat(s, '-', 2)
		}
	default:
		panic(fmt.Errorf("unexpected djot token: %v(%d)", tokenType, tokenType))
	}
	return
}
