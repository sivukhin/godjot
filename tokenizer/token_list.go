package tokenizer

import (
	"fmt"
	"strings"
)

type TokenList[T ~int] []Token[T]

func DefaultType[T ~int]() T {
	var defaultType T
	return defaultType
}

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
	l.FillUntil(token.Start, DefaultType[T]())
	*l = append(*l, token)
}

func (l *TokenList[T]) FillUntil(position int, tokenType T) {
	last := l.LastOrDefault()
	if len(*l) > 0 && last.End < position {
		*l = append(*l, Token[T]{Type: tokenType, Start: last.End, End: position})
	}
}

func (l TokenList[T]) GoString() string {
	tokens := make([]string, len(l))
	for i, token := range l {
		tokens[i] = fmt.Sprintf("%v", token)
	}
	return strings.Join(tokens, ",")
}
