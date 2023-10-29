package tokenizer

type Token[T comparable] struct {
	Type       T
	JumpToPair int
	Start, End int
	Attributes map[string]string
}

func (t Token[T]) Length() int { return t.End - t.Start }
