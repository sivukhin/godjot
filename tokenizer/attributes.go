package tokenizer

const (
	DjotAttributeClassKey = "class"
	DjotAttributeIdKey    = "id"
)

type DjotAttribute struct{ Key, Value string }

func MatchQuotedString(r TextReader, s ReaderState) (value []byte, next ReaderState) {
	var rawBytesMask = NewByteMask([]byte("\\\"")).Negate()

	next = r.Token(s, "\"")
	if !next.Matched() {
		return
	}
	start := next
	for {
		next = r.MaskRepeat(next, rawBytesMask, 0)
		value = append(value, r[start:next]...)
		start = next
		if endString := r.Token(next, "\""); endString.Matched() {
			return value, endString
		} else if escape := r.Token(next, "\\"); escape.Matched() {
			if r.Empty(escape) {
				return nil, Unmatched
			}
			value = append(value, r[escape])
			start = escape + 1
			next = escape + 1
		} else {
			return nil, Unmatched
		}
	}
}

func MatchDjotAttribute(r TextReader, s ReaderState) (attributes map[string]string, next ReaderState) {
	attributes = make(map[string]string, 0)
	next = r.Token(s, "{")
	if !next.Matched() {
		return
	}
	comment := false
	for {
		next = r.MaskRepeat(next, SpaceByteMask, 0)
		if r.Empty(next) {
			next = Unmatched
			return
		}
		if commentStart := r.Token(next, "%"); commentStart.Matched() {
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
		if attributeEnd := r.Token(next, "}"); attributeEnd.Matched() {
			return attributes, attributeEnd
		}
		startKey := next
		if classToken := r.Token(next, "."); classToken.Matched() {
			next = r.MaskRepeat(classToken, AttributeTokenMask, 1)
		} else if idToken := r.Token(next, "#"); idToken.Matched() {
			next = r.MaskRepeat(idToken, AttributeTokenMask, 1)
		} else {
			next = r.MaskRepeat(next, AttributeTokenMask, 1)
		}
		if !next.Matched() {
			return
		}
		key := r.Select(startKey, next)
		if key[0] == '.' {
			if class, hasClass := attributes[DjotAttributeClassKey]; hasClass {
				attributes[DjotAttributeClassKey] = class + " " + key[1:]
			} else {
				attributes[DjotAttributeClassKey] = key[1:]
			}
			continue
		}
		if key[0] == '#' {
			attributes[DjotAttributeIdKey] = key[1:]
			continue
		}

		next = r.Token(next, "=")
		if !next.Matched() {
			return
		}
		startValue := next
		if value, quoteEnd := MatchQuotedString(r, next); quoteEnd.Matched() {
			attributes[key] = string(value)
			next = quoteEnd
		} else {
			next = r.MaskRepeat(next, AttributeTokenMask, 1)
			if !next.Matched() {
				return
			}
			attributes[key] = r.Select(startValue, next)
		}
	}
}
