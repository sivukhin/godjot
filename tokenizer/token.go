package tokenizer

const Open = 1

type Token[T comparable] struct {
	Type       T
	JumpToPair int
	Start, End int
	Attributes *Attributes
}

func (t Token[T]) Length() int { return t.End - t.Start }

func (t Token[T]) IsDefault() bool {
	var defaultType T
	return defaultType == t.Type
}

func (t Token[T]) Bytes(document []byte) []byte { return document[t.Start:t.End] }
func (t Token[T]) PrefixLength(document []byte, b byte) int {
	bytes := t.Bytes(document)
	i := 0
	for i < len(bytes) && bytes[i] == b {
		i++
	}
	return i
}
