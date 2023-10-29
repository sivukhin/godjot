package tokenizer

import (
	"fmt"
)

const Open = 1

type TokenLevel[T ~int] []Token[T]

func (l TokenLevel[T]) First() *Token[T] {
	if len(l) == 0 {
		return &Token[T]{}
	}
	return &l[0]
}

func (l TokenLevel[T]) Last() *Token[T] {
	if len(l) == 0 {
		return &Token[T]{}
	}
	return &l[len(l)-1]
}

type Tokenizer[T ~int] struct {
	Levels     []TokenLevel[T]
	TypeLevels map[T][]int
}

func NewTokenizer[T ~int]() Tokenizer[T] {
	return Tokenizer[T]{Levels: nil, TypeLevels: make(map[T][]int)}
}

func (b *Tokenizer[T]) LastLevel() *TokenLevel[T] {
	if len(b.Levels) == 0 {
		panic(fmt.Errorf("unable to get last level of empty builder"))
	}
	return &b.Levels[len(b.Levels)-1]
}

func (b *Tokenizer[T]) PopCommit() {
	if len(b.Levels) <= 1 {
		panic(fmt.Errorf("unable to pop from the builder with only %v levels", len(b.Levels)))
	}
	lastLevel := b.LastLevel()
	b.Levels = b.Levels[0 : len(b.Levels)-1]

	var firstPosition, lastPosition int
	for i, token := range *lastLevel {
		var defaultType T
		if defaultType == token.Type {
			continue
		}
		b.addToken(token, true)
		if i == 0 {
			firstPosition = len(*b.LastLevel()) - 1
		}
		if i == len(*lastLevel)-1 {
			lastPosition = len(*b.LastLevel()) - 1
		}
	}
	if lastLevel.First().Type^Open == lastLevel.Last().Type {
		(*b.LastLevel())[firstPosition].JumpToPair = lastPosition - firstPosition
		(*b.LastLevel())[lastPosition].JumpToPair = -(lastPosition - firstPosition)
	}
	if typeLevels := b.TypeLevels[lastLevel.First().Type]; len(typeLevels) > 0 {
		b.TypeLevels[lastLevel.First().Type] = typeLevels[0 : len(typeLevels)-1]
	}
}

func (b *Tokenizer[T]) PopForget() {
	if len(b.Levels) <= 1 {
		panic(fmt.Errorf("unable to pop from the builder with only %v levels", len(b.Levels)))
	}
	if typeLevels := b.TypeLevels[b.LastLevel().First().Type]; len(typeLevels) > 0 {
		b.TypeLevels[b.LastLevel().First().Type] = typeLevels[0 : len(typeLevels)-1]
	}
	lastLevel := b.LastLevel()
	b.Levels = b.Levels[0 : len(b.Levels)-1]
	for _, token := range (*lastLevel)[1:] {
		b.addToken(token, true)
	}
}

func (b *Tokenizer[T]) PopForgetUntil(tokenType T) bool {
	levels := b.TypeLevels[tokenType]
	if len(levels) == 0 {
		return false
	}
	lastLevel := levels[len(levels)-1]
	for len(b.Levels) > lastLevel+1 {
		b.PopForget()
	}
	return true
}

func (b *Tokenizer[T]) addToken(token Token[T], insertEmpty bool) {
	if len(b.Levels) == 0 {
		panic(fmt.Errorf("unable to add raw token with no levels: token=%v", token))
	}
	lastLevel := b.LastLevel()
	if last := b.LastLevel().Last(); insertEmpty && last.End < token.Start {
		*lastLevel = append(*lastLevel, Token[T]{Start: last.End, End: token.Start})
	}
	*lastLevel = append(*lastLevel, token)
}

func (b *Tokenizer[T]) OpenLevelAt(tokenType T, start, end int, attributes ...map[string]string) {
	if len(b.Levels) > 0 {
		if last := b.LastLevel().Last(); last.End < start {
			b.addToken(Token[T]{Start: last.End, End: start}, false)
		}
	}
	b.TypeLevels[tokenType] = append(b.TypeLevels[tokenType], len(b.Levels))
	b.Levels = append(b.Levels, nil)
	b.addToken(Token[T]{Type: tokenType, Start: start, End: end, Attributes: getOptional(attributes...)}, false)
}

func (b *Tokenizer[T]) CloseLevelAt(tokenType T, start, end int) {
	b.AddLengthToken(tokenType, start, end)
	b.PopCommit()
}

func (b *Tokenizer[T]) AddLengthToken(tokenType T, start, end int, attributes ...map[string]string) {
	b.addToken(Token[T]{Type: tokenType, Start: start, End: end, Attributes: getOptional(attributes...)}, true)
}

func getOptional[T any](value ...T) (result T) {
	if len(value) > 0 {
		result = value[0]
	}
	return
}
