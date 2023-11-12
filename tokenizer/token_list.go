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
	l.FillUntil(token.Start)
	*l = append(*l, token)
}

func (l *TokenList[T]) FillUntil(position int, tokenTypeOpt ...T) {
	var tokenType T
	if len(tokenTypeOpt) > 0 {
		tokenType = tokenTypeOpt[0]
	}
	last := l.LastOrDefault()
	if len(*l) > 0 && last.End < position {
		*l = append(*l, Token[T]{Type: tokenType, Start: last.End, End: position})
	}
}
