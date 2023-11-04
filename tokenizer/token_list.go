package tokenizer

type TokenList[T ~int] []Token[T]

func (l *TokenList[T]) FirstOrDefault() *Token[T] {
	if l == nil || len(*l) == 0 {
		return &Token[T]{}
	}
	return &(*l)[0]
}

func (l *TokenList[T]) LastOrDefault() *Token[T] {
	if l == nil || len(*l) == 0 {
		return &Token[T]{}
	}
	return &(*l)[len(*l)-1]
}

func (l *TokenList[T]) Push(token Token[T]) {
	last := l.LastOrDefault()
	if len(*l) > 0 && last.End < token.Start {
		*l = append(*l, Token[T]{Start: last.End, End: token.Start})
	}
	*l = append(*l, token)
}
