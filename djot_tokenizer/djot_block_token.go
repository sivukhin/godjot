package djot_tokenizer

import (
	"bytes"

	"git.sr.ht/~ser/godjot/v2/tokenizer"
)

var (
	NotSpaceNewLineByteMask = tokenizer.SpaceNewLineByteMask.Negate()
	NotBracketByteMask      = tokenizer.NewByteMask([]byte("]")).Negate()
	ThematicBreakByteMask   = tokenizer.NewByteMask([]byte(" \t\n*-"))

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
	initialState tokenizer.ReaderState,
	tokenType DjotToken,
) (tokenizer.Token[DjotToken], tokenizer.ReaderState, bool) {
	fail := func() (tokenizer.Token[DjotToken], tokenizer.ReaderState, bool) {
		return tokenizer.Token[DjotToken]{}, initialState, false
	}
	success := func(token tokenizer.Token[DjotToken], state tokenizer.ReaderState) (tokenizer.Token[DjotToken], tokenizer.ReaderState, bool) {
		return token, state, true
	}

	var ok bool

	initialState, ok = r.MaskRepeat(initialState, tokenizer.SpaceByteMask, 0)
	tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

	next := initialState // initialize next variable to initialState in order to use next, ok = r.Matcher(next, ...) structure whenever it's possible

	switch tokenType {
	case HeadingBlock:
		if next, ok = r.ByteRepeat(next, '#', 1); !ok {
			return fail()
		}
		if next, ok = r.Mask(next, tokenizer.SpaceByteMask); !ok {
			return fail()
		}
		return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: next}, next)
	case QuoteBlock:
		if next, ok = r.Token(next, ">"); !ok {
			return fail()
		}
		if next, ok = r.Mask(next, tokenizer.SpaceNewLineByteMask); !ok {
			return fail()
		}
		return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: next}, next)
	case DivBlock, CodeBlock:
		var symbol byte
		var attributeKey string
		switch tokenType {
		case DivBlock:
			symbol, attributeKey = ':', DjotAttributeClassKey
		case CodeBlock:
			symbol, attributeKey = '`', CodeLangKey
		}

		if next, ok = r.ByteRepeat(next, symbol, 3); !ok {
			return fail()
		}

		repeat := next

		next, ok = r.MaskRepeat(next, tokenizer.SpaceByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
		if end, ok := r.EmptyOrWhiteSpace(next); ok {
			return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: end}, end)
		}

		metaStart := next
		next, ok = r.MaskRepeat(next, NotSpaceNewLineByteMask, 1)
		// usually MaskRepeat must match because !r.IsEmpty(next) and next symbol is not in SpaceByteMask but in some broken cases this can fail (see panic1 test)
		if !ok {
			return fail()
		}
		metaEnd := next

		if next, ok = r.EmptyOrWhiteSpace(next); !ok {
			return fail()
		}

		token := tokenizer.Token[DjotToken]{
			Type:  tokenType,
			Start: initialState,
			End:   repeat,
			Attributes: tokenizer.NewAttributes(tokenizer.AttributeEntry{
				Key:   attributeKey,
				Value: r.Select(metaStart, metaEnd),
			}),
		}
		return success(token, next)
	case ReferenceDefBlock, FootnoteDefBlock:
		var blockToken string
		switch tokenType {
		case ReferenceDefBlock:
			blockToken = "["
		case FootnoteDefBlock:
			blockToken = "[^"
		}
		if next, ok = r.Token(next, blockToken); !ok {
			return fail()
		}
		startKey := next
		next, ok = r.MaskRepeat(next, NotBracketByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
		endKey := next

		if next, ok = r.Token(next, "]:"); !ok {
			return fail()
		}
		token := tokenizer.Token[DjotToken]{
			Type:  tokenType,
			Start: initialState,
			End:   next,
			Attributes: tokenizer.NewAttributes(tokenizer.AttributeEntry{
				Key:   ReferenceKey,
				Value: r.Select(startKey, endKey),
			}),
		}
		return success(token, next)
	case ThematicBreakToken:
		next, ok = r.MaskRepeat(next, ThematicBreakByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
		if !r.IsEmpty(next) {
			return fail()
		}
		// three or more * or - characters required for thematic break
		if bytes.Count(r[initialState:next], []byte{'*'}) < 3 && bytes.Count(r[initialState:next], []byte{'-'}) < 3 {
			return fail()
		}
		return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: next}, next)
	case ListItemBlock:
		for _, simpleToken := range [...]string{"- [ ] ", "- [x] ", "- [X] ", "+ ", "* ", "- ", ": "} {
			if simple, ok := r.Token(next, simpleToken); ok {
				return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: simple}, simple)
			}
		}

		for _, complexTokenMask := range [...]tokenizer.ByteMask{DigitByteMask, LowerAlphaByteMask, UpperAlphaByteMask} {
			// consider three valid cases: (MASK) | MASK) | MASK.
			complexNext := next
			parenOpen, parenOk := r.Token(next, "(")
			if parenOk {
				complexNext = parenOpen
			}
			if complexNext, ok = r.MaskRepeat(complexNext, complexTokenMask, 1); !ok {
				continue
			}
			if ending, ok := r.Token(complexNext, ") "); ok {
				return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: ending}, ending)
			} else if ending, ok := r.Token(complexNext, ". "); ok && !parenOk {
				return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: ending}, ending)
			}
		}
		return fail()
	case PipeTableBlock:
		if next >= len(r) || r[next] != '|' {
			return fail()
		}
		last := len(r) - 1
		for last > next && r.HasMask(last, tokenizer.SpaceNewLineByteMask) {
			last--
		}
		if r[last] != '|' {
			return fail()
		}
		return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: initialState}, initialState)
	case ParagraphBlock:
		if r.IsEmpty(next) {
			return fail()
		}
		return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: next}, next)
	case PipeTableCaptionBlock:
		if next, ok = r.Token(next, "^ "); !ok {
			return fail()
		}
		return success(tokenizer.Token[DjotToken]{Type: tokenType, Start: initialState, End: next}, next)
	default:
		tokenizer.Assertf(false, "unexpected djot block token type: %v", tokenType)
		return fail()
	}
}
