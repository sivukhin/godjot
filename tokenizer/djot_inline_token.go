package tokenizer

import (
	"fmt"
	"unicode"
)

var (
	DollarByteMask   = NewByteMask([]byte("$"))
	BacktickByteMask = NewByteMask([]byte("`"))
)

func MatchInlineToken(
	r TextReader,
	s ReaderState,
	tokenType DjotToken,
) (next ReaderState) {
	next = Unmatched

	switch tokenType {
	case Span:
		next = r.Token(s, "[")
	case Span ^ Open:
		next = r.Token(s, "]")
	case LinkUrl:
		next = r.Token(s, "(")
	case LinkUrl ^ Open:
		next = r.Token(s, ")")
	case LinkReference:
		next = r.Token(s, "[")
	case LinkReference ^ Open:
		next = r.Token(s, "]")
	case Autolink:
		next = r.Token(s, "<")
	case Autolink ^ Open:
		next = r.Token(s, ">")
	case Verbatim:
		next = r.MaskRepeat(s, DollarByteMask, 0)
		if !next.Matched() {
			break
		}
		if next-s > 2 {
			break
		}
		next = r.MaskRepeat(next, BacktickByteMask, 1)
	case Verbatim ^ Open:
		next = r.MaskRepeat(s, BacktickByteMask, 1)
	case Emphasis:
		next = r.Token(s, "{_")
		if next.Matched() {
			break
		}
		next = r.Token(s, "_")
		if r.HasMask(next, SpaceNewLineByteMask) {
			next = Unmatched
		}
	case Emphasis ^ Open:
		next = r.Token(s, "_}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "_")
		if r.HasMask(s-1, SpaceNewLineByteMask) {
			next = Unmatched
		}
	case Strong:
		next = r.Token(s, "{*")
		if next.Matched() {
			break
		}
		next = r.Token(s, "*")
		if r.HasMask(next, SpaceNewLineByteMask) {
			next = Unmatched
		}
	case Strong ^ Open:
		next = r.Token(s, "*}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "*")
		if r.HasMask(s-1, SpaceNewLineByteMask) {
			next = Unmatched
		}
	case Highlighted:
		next = r.Token(s, "{=")
	case Highlighted ^ Open:
		next = r.Token(s, "=}")
	case Superscript:
		next = r.Token(s, "{^")
		if next.Matched() {
			break
		}
		next = r.Token(s, "^")
	case Superscript ^ Open:
		next = r.Token(s, "^}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "^")
	case Subscript:
		next = r.Token(s, "{~")
		if next.Matched() {
			break
		}
		next = r.Token(s, "~")
	case Subscript ^ Open:
		next = r.Token(s, "~}")
		if next.Matched() {
			break
		}
		next = r.Token(s, "~")
	case Insert:
		next = r.Token(s, "{+")
	case Insert ^ Open:
		next = r.Token(s, "+}")
	case Delete:
		next = r.Token(s, "{-")
	case Delete ^ Open:
		next = r.Token(s, "-}")
	case FootnoteReference:
		next = r.Token(s, "[^")
	case FootnoteReference ^ Open:
		next = r.Token(s, "]")
	case Escaped:
		next = r.Token(s, "\\")
		if !next.Matched() {
			break
		}
		if r.Empty(next) {
			next = Unmatched
			break
		}
		if symbol := r[next]; unicode.IsPunct(rune(symbol)) {
			next++
		} else {
			next = r.MaskRepeat(next, SpaceByteMask, 0)
			next = r.Token(next, "\n")
		}
	case RawFormat:
		next = r.Token(s, "{=")
	case RawFormat ^ Open:
		next = r.Token(s, "}")
	case Symbols, Symbols ^ Open:
		next = r.Token(s, ":")
	default:
		panic(fmt.Errorf("unexpected djot token: %v(%d)", tokenType, tokenType))
	}
	return
}
