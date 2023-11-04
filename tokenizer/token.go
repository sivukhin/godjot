package tokenizer

const Open = 1

type Token[T comparable] struct {
	Type       T
	JumpToPair int
	Start, End int
	Attributes map[string]string
}

func (t Token[T]) Length() int { return t.End - t.Start }

func (t Token[T]) IsDefault() bool {
	var defaultType T
	return defaultType == t.Type
}
