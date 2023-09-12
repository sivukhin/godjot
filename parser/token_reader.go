package parser

import "github.com/sivukhin/godjot/tokenizer"

type TokenReader[T comparable] struct {
	Text     []byte
	Tokens   []tokenizer.Token[T]
	Position int
}

func (r TokenReader[T]) Empty() bool { return r.Position >= len(r.Tokens) }

func (r TokenReader[T]) Peek(d int) (tokenizer.Token[T], bool) {
	if r.Position+d < 0 {
		return tokenizer.Token[T]{}, false
	}
	if r.Position+d >= len(r.Tokens) {
		return tokenizer.Token[T]{}, false
	}
	return r.Tokens[r.Position+d], true
}

func (r TokenReader[T]) MatchToken(tokenType T, next *TokenReader[T]) (tokenizer.Token[T], bool) {
	if current, ok := r.Peek(0); ok && current.Type == tokenType {
		*next = TokenReader[T]{Text: r.Text, Tokens: r.Tokens, Position: r.Position + 1}
		return current, true
	}
	return tokenizer.Token[T]{}, false
}
