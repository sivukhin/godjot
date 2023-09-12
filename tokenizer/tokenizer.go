package tokenizer

import (
	"fmt"
)

type Token[T comparable] struct {
	Type       T
	JumpToPair int
	Start, End int
}

type TokenLevel[T ~int] []Token[T]

type Tokenizer[T ~int] struct {
	Levels     []TokenLevel[T]
	TypeLevels map[T][]int
}

func (b *Tokenizer[T]) LastLevel() Token[T] {
	if len(b.Levels) == 0 {
		panic(fmt.Errorf("unable to get last level of empty builder"))
	}
	return b.Levels[len(b.Levels)-1][0]
}

func (b *Tokenizer[T]) Pop(skipFirstTokens int) {
	if len(b.Levels) <= 1 {
		panic(fmt.Errorf("unable to pop from the builder with only %v levels", len(b.Levels)))
	}
	last := b.Levels[len(b.Levels)-1]
	b.Levels = b.Levels[0 : len(b.Levels)-1]
	if last[0].Type^Close == last[len(last)-1].Type {
		last[0].JumpToPair = len(last) - 1
		last[len(last)-1].JumpToPair = -len(last) + 1
	}
	for _, token := range last[skipFirstTokens:] {
		b.AddToken(token)
	}
	if typeLevels := b.TypeLevels[last[0].Type]; len(typeLevels) > 0 {
		b.TypeLevels[last[0].Type] = typeLevels[0 : len(typeLevels)-1]
	}
}

func (b *Tokenizer[T]) PopCommit() { b.Pop(0) }
func (b *Tokenizer[T]) PopForget() { b.Pop(1) }

func (b *Tokenizer[T]) PopForgetUntil(tokenType T) bool {
	levels := b.TypeLevels[tokenType]
	if len(levels) == 0 {
		return false
	}
	lastLevel := levels[len(levels)-1]
	for i := lastLevel; i+1 < len(b.Levels); i++ {
		b.PopForget()
	}
	return true
}

func (b *Tokenizer[T]) AddToken(token Token[T]) {
	if len(b.Levels) == 0 {
		panic(fmt.Errorf("unable to add raw token with no levels: token=%v", token))
	}
	if len(b.Levels[len(b.Levels)-1]) > 0 {
		last := &b.Levels[len(b.Levels)-1][len(b.Levels[len(b.Levels)-1])-1]
		if last.End < token.Start {
			var raw T
			b.Levels[len(b.Levels)-1] = append(b.Levels[len(b.Levels)-1], Token[T]{Type: raw, Start: last.End, End: token.Start})
		}
	}
	b.Levels[len(b.Levels)-1] = append(b.Levels[len(b.Levels)-1], token)
}
func (b *Tokenizer[T]) OpenLevelAt(tokenType T, start int, end int) {
	b.TypeLevels[tokenType] = append(b.TypeLevels[tokenType], len(b.Levels))
	b.Levels = append(b.Levels, nil)
	b.AddLengthToken(tokenType, start, end)
}
func (b *Tokenizer[T]) CloseLevelAt(tokenType T, start int, end int) {
	b.AddLengthToken(tokenType, start, end)
	b.PopCommit()
}

func (b *Tokenizer[T]) AddLengthToken(tokenType T, start, end int) {
	b.AddToken(Token[T]{Type: tokenType, Start: start, End: end})
}
