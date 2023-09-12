package tokenizer

import "fmt"

func matchDjotTokenPair(reader TextReader, next *TextReader, builder *Tokenizer[DjotToken], token DjotToken) bool {
	if MatchDjotToken(reader, token|Close, next) && builder.PopForgetUntil(token) {
		builder.CloseLevelAt(token|Close, reader.Position, next.Position)
		return true
	}
	if MatchDjotToken(reader, token, next) {
		builder.OpenLevelAt(token, reader.Position, next.Position)
		return true
	}
	return false
}

func djotToken(reader TextReader, builder *Tokenizer[DjotToken]) TextReader {
	var next TextReader
	lastLevel := builder.LastLevel()
	if lastLevel.Type != Verbatim && MatchDjotToken(reader, Escaped, &next) {
		builder.AddLengthToken(Escaped, reader.Position, next.Position)
		return next
	}

	switch lastLevel.Type {
	case Verbatim:
		if MatchDjotToken(reader, Verbatim|Close, &next) && next.Position-reader.Position == lastLevel.End-lastLevel.Start {
			builder.CloseLevelAt(Verbatim|Close, reader.Position, next.Position)
			return next
		}
	case String:
		if matchDjotTokenPair(reader, &next, builder, String) {
			return next
		}
	case Attribute:
		if MatchDjotToken(reader, Attribute|Close, &next) {
			builder.CloseLevelAt(Attribute|Close, reader.Position, next.Position)
			return next
		}
		if matchDjotTokenPair(reader, &next, builder, String) {
			return next
		}
	default:
		for token := firstSimpleToken; token <= lastSimpleToken; token += 2 {
			if matchDjotTokenPair(reader, &next, builder, token) {
				return next
			}
		}
	}
	return reader.MatchAny()
}

func DjotTokens(text []byte) []Token[DjotToken] {
	reader := TextReader{Text: text}

	builder := Tokenizer[DjotToken]{Levels: []TokenLevel[DjotToken]{{}}, TypeLevels: make(map[DjotToken][]int)}
	builder.OpenLevelAt(Doc, 0, 0)
	for !reader.Empty() {
		reader = djotToken(reader, &builder)
	}
	if !builder.PopForgetUntil(Doc) {
		panic(fmt.Errorf("unable to find root Doc element"))
	}
	builder.CloseLevelAt(Doc|Close, len(text), len(text))
	return builder.Levels[0]
}
