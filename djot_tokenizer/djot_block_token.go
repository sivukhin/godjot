package djot_tokenizer

import (
	"bytes"

	"github.com/sivukhin/godjot/tokenizer"
)

var (
	NotSpaceByteMask      = tokenizer.SpaceNewLineByteMask.Negate()
	NotBracketByteMask    = tokenizer.NewByteMask([]byte("]")).Negate()
	ThematicBreakByteMask = tokenizer.NewByteMask([]byte(" \t\n*-"))

	DigitByteMask      = tokenizer.NewByteMask([]byte("0123456789"))
	LowerAlphaByteMask = tokenizer.NewByteMask([]byte("abcdefghijklmnopqrstuvwxyz"))
	UpperAlphaByteMask = tokenizer.NewByteMask([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	AttributeTokenMask = tokenizer.Union(
		DigitByteMask,
		LowerAlphaByteMask,
		UpperAlphaByteMask,
		tokenizer.NewByteMask([]byte(`-_:`)),
	)
)

func MatchBlockToken(
	r tokenizer.TextReader,
	s tokenizer.ReaderState,
	tokenType DjotToken,
) (token tokenizer.Token[DjotToken], next tokenizer.ReaderState) {
	next = tokenizer.Unmatched

	s = r.MaskRepeat(s, tokenizer.SpaceByteMask, 0)
	switch tokenType {
	case HeadingBlock:
		if next = r.ByteRepeat(s, '#', 1); !next.Matched() {
			return
		}
		if next = r.Mask(next, tokenizer.SpaceByteMask); !next.Matched() {
			return
		}
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case QuoteBlock:
		if next = r.Token(s, ">"); !next.Matched() {
			return
		}
		if next = r.Mask(next, tokenizer.SpaceNewLineByteMask); !next.Matched() {
			return
		}
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case DivBlock, CodeBlock:
		var symbol byte
		var attributeKey string
		switch tokenType {
		case DivBlock:
			symbol, attributeKey = ':', DivClassKey
		case CodeBlock:
			symbol, attributeKey = '`', CodeLangKey
		}

		if next = r.ByteRepeat(s, symbol, 3); !next.Matched() {
			return
		}

		next = r.MaskRepeat(next, tokenizer.SpaceByteMask, 0) // will always match because count = 0
		if r.EmptyOrWhiteSpace(next).Matched() {
			token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
			return
		}

		metaStart := next
		next = r.MaskRepeat(next, NotSpaceByteMask, 1) // will always match because !r.Empty(next) and next symbol is not in SpaceByteMask
		metaEnd := next

		next = r.EmptyOrWhiteSpace(next)
		if !next.Matched() {
			return
		}

		attributes := tokenizer.Attributes{}
		token = tokenizer.Token[DjotToken]{
			Type:       tokenType,
			Start:      int(s),
			End:        int(next),
			Attributes: attributes.Set(attributeKey, r.Select(metaStart, metaEnd)),
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
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case ThematicBreakToken:
		next = r.MaskRepeat(s, ThematicBreakByteMask, 0)
		if !r.Empty(next) {
			next = tokenizer.Unmatched
			return
		}
		if bytes.Count(r[s:next], []byte("*")) < 3 && bytes.Count(r[s:next], []byte("-")) < 3 {
			next = tokenizer.Unmatched
			return
		}
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case ListItemBlock:
		for _, simpleToken := range []string{"+ ", "* ", "- ", ": ", "- [ ] ", "- [x] ", "- [X] "} {
			next = r.Token(s, simpleToken)
			if next.Matched() {
				token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
				return
			}
		}
		for _, complexToken := range []tokenizer.ByteMask{DigitByteMask, LowerAlphaByteMask, UpperAlphaByteMask} {
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
				} else if dotEnding.Matched() {
					next = dotEnding
					break
				} else {
					next = tokenizer.Unmatched
				}
			}
		}
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(next)}
	case PipeTableBlock:
		// todo (sivukhin, 2023-10-28): check that line ends with another pipe
		next = r.Token(s, "|")
		if !next.Matched() {
			return
		}
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(s)}
	case ParagraphBlock:
		if r.Empty(s) {
			return
		}
		next = s
		token = tokenizer.Token[DjotToken]{Type: tokenType, Start: int(s), End: int(s)}
	}
	return
}
