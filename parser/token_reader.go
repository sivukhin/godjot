package parser

import "github.com/sivukhin/godjot/tokenizer"

type TokenPosition = int
type TokenReader[T comparable] struct {
	Text   []byte
	Tokens []tokenizer.Token[T]
}

func (r *TokenReader[T]) Span(start, end int) *TokenReader[T] {
	return &TokenReader[T]{Text: r.Text, Tokens: r.Tokens[start:end]}
}
func (r *TokenReader[T]) Empty(current TokenPosition) bool { return current >= len(r.Tokens) }
func (r *TokenReader[T]) Peek(current TokenPosition, d int) (tokenizer.Token[T], bool) {
	if current+d < 0 {
		return tokenizer.Token[T]{}, false
	}
	if current+d >= len(r.Tokens) {
		return tokenizer.Token[T]{}, false
	}
	return r.Tokens[current+d], true
}
func (r *TokenReader[T]) MatchToken(current TokenPosition, next *TokenPosition, tokenType T) (tokenizer.Token[T], bool) {
	if peek, ok := r.Peek(current, 0); ok && peek.Type == tokenType {
		*next = current + 1
		return peek, true
	}
	return tokenizer.Token[T]{}, false
}
