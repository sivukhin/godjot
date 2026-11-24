package djot_tokenizer

import "git.sr.ht/~ser/godjot/v2/tokenizer"

const (
	DjotAttributeClassKey = "class"
	DjotAttributeIdKey    = "id"
)

func MatchQuotedString(r tokenizer.TextReader, s tokenizer.ReaderState) ([]byte, tokenizer.ReaderState, bool) {
	fail := func() ([]byte, tokenizer.ReaderState, bool) { return nil, 0, false }

	var rawBytesMask = tokenizer.NewByteMask([]byte("\\\"")).Negate()

	next, ok := r.Token(s, "\"")
	if !ok {
		return fail()
	}

	var value []byte
	start := next
	for {
		next, ok = r.MaskRepeat(next, rawBytesMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

		value = append(value, r[start:next]...)
		if endString, ok := r.Token(next, "\""); ok {
			return value, endString, true
		} else if escape, ok := r.Token(next, "\\"); ok {
			if r.IsEmpty(escape) {
				return fail()
			}
			value = append(value, r[escape])
			start = escape + 1
			next = escape + 1
		} else {
			return fail()
		}
	}
}

func MatchDjotAttribute(r tokenizer.TextReader, s tokenizer.ReaderState) (tokenizer.Attributes, tokenizer.ReaderState, bool) {
	fail := func() (tokenizer.Attributes, tokenizer.ReaderState, bool) { return tokenizer.Attributes{}, 0, false }

	next, ok := r.Token(s, "{")
	if !ok {
		return fail()
	}
	var attributes tokenizer.Attributes
	comment := false
	for {
		next, ok = r.MaskRepeat(next, tokenizer.SpaceNewLineByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

		if r.IsEmpty(next) {
			return fail()
		}
		if commentStart, ok := r.Token(next, "%"); ok {
			if comment {
				comment = false
			} else {
				comment = true
			}
			next = commentStart
			continue
		}
		if comment {
			next++
			continue
		}
		if attributeEnd, ok := r.Token(next, "}"); ok {
			return attributes, attributeEnd, true
		}
		if classToken, ok := r.Token(next, "."); ok {
			if next, ok = r.MaskRepeat(classToken, AttributeTokenMask, 1); !ok {
				return fail()
			}
			className := r.Select(classToken, next)
			attributes.Append(DjotAttributeClassKey, className)
			continue
		} else if idToken, ok := r.Token(next, "#"); ok {
			if next, ok = r.MaskRepeat(idToken, AttributeTokenMask, 1); !ok {
				return fail()
			}
			attributes.Set(DjotAttributeIdKey, r.Select(idToken, next))
			continue
		}
		startKey := next
		if next, ok = r.MaskRepeat(next, AttributeTokenMask, 1); !ok {
			return fail()
		}
		endKey := next

		if next, ok = r.Token(next, "="); !ok {
			return fail()
		}

		startValue := next
		if value, quoteEnd, ok := MatchQuotedString(r, next); ok {
			attributes.Set(r.Select(startKey, endKey), string(value))
			next = quoteEnd
		} else {
			if next, ok = r.MaskRepeat(next, AttributeTokenMask, 1); !ok {
				return fail()
			}
			attributes.Set(r.Select(startKey, endKey), r.Select(startValue, next))
		}
	}
}
