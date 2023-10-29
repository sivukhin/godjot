package tokenizer

import (
	"bytes"
)

var (
	NotSpaceByteMask      = SpaceByteMask.Negate()
	NotBracketByteMask    = NewByteMask([]byte("]")).Negate()
	ThematicBreakByteMask = NewByteMask([]byte(" \t*-"))

	DigitByteMask      = NewByteMask([]byte("0123456789"))
	LowerAlphaByteMask = NewByteMask([]byte("abcdefghijklmnopqrstuvwxyz"))
	UpperAlphaByteMask = NewByteMask([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	AttributeTokenMask = Union(DigitByteMask, LowerAlphaByteMask, UpperAlphaByteMask, NewByteMask([]byte(`-_:`)))
)

func MatchBlockToken(
	r TextReader,
	s ReaderState,
	tokenType DjotToken,
) (token Token[DjotToken], next ReaderState) {
	next = Unmatched

	s = r.MaskRepeat(s, SpaceByteMask, 0)
	switch tokenType {
	case HeadingBlock:
		if next = r.ByteRepeat(s, '#', 3); !next.Matched() {
			return
		}
		headingEnd := next
		if next = r.Mask(next, SpaceByteMask); !next.Matched() {
			return
		}
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(headingEnd)}
	case QuoteBlock:
		if next = r.Token(s, ">"); !next.Matched() {
			return
		}
		if next = r.Mask(next, SpaceByteMask); !next.Matched() {
			return
		}
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(s + 1)}
	case DivBlock, CodeBlock:
		var symbol byte
		switch tokenType {
		case DivBlock:
			symbol = ':'
		case CodeBlock:
			symbol = '`'
		}

		if next = r.ByteRepeat(s, symbol, 3); !next.Matched() {
			return
		}
		tokenEnd := next

		next = r.MaskRepeat(next, SpaceByteMask, 0) // will always match because count = 0
		if r.Empty(next) {
			token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(tokenEnd)}
			return
		}

		metaStart := next
		next = r.MaskRepeat(next, NotSpaceByteMask, 1) // will always match because !r.Empty(next) and next symbol is not in SpaceByteMask
		metaEnd := next

		next = r.EmptyOrWhiteSpace(next)
		if !next.Matched() {
			return
		}

		token = Token[DjotToken]{
			Type:       tokenType,
			Start:      int(s),
			End:        int(tokenEnd),
			Attributes: map[string]string{"Meta": r.Select(metaStart, metaEnd)},
		}
	case ReferenceDefBlock, FootnoteDefBlock:
		var blockToken string
		switch tokenType {
		case ReferenceDefBlock:
			blockToken = "["
		case FootnoteDefBlock:
			blockToken = "[^"
		}
		next = r.Token(s, blockToken)
		if !next.Matched() {
			return
		}
		next = r.MaskRepeat(next, NotBracketByteMask, 0) // will always match because count = 0

		next = r.Token(next, "]:")
		if !next.Matched() {
			return
		}
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case ThematicBreakToken:
		next = r.MaskRepeat(s, ThematicBreakByteMask, 0)
		if !r.Empty(next) {
			next = Unmatched
			return
		}
		if bytes.Count(r[s:next], []byte("-*")) < 3 {
			next = Unmatched
			return
		}
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case ListItemBlock:
		for _, simpleToken := range []string{"+ ", "* ", "- ", ": ", "- [ ] ", "- [x] ", "- [X] "} {
			next = r.Token(s, simpleToken)
			if next.Matched() {
				return
			}
		}
		for _, complexToken := range []ByteMask{DigitByteMask, LowerAlphaByteMask, UpperAlphaByteMask} {
			next = r.Token(s, "(")
			if next.Matched() {
				next = r.MaskRepeat(next, complexToken, 1)
				if !next.Matched() {
					continue
				}
				next = r.Token(next, ") ")
				if next.Matched() {
					break
				}
			} else {
				next = r.MaskRepeat(s, complexToken, 1)
				if !next.Matched() {
					continue
				}
				parEnding := r.Token(next, ") ")
				dotEnding := r.Token(next, ". ")
				if parEnding.Matched() {
					next = parEnding
					break
				}
				if dotEnding.Matched() {
					next = dotEnding
					break
				}
			}
		}
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case PipeTableBlock:
		// todo (sivukhin, 2023-10-28): check that line ends with another pipe
		next = r.Token(s, "|")
		if !next.Matched() {
			return
		}
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(s)}
	case ParagraphToken:
		if r.Empty(s) {
			return
		}
		next = s
		token = Token[DjotToken]{Type: tokenType, Start: int(s), End: int(s)}
	}
	return
}
