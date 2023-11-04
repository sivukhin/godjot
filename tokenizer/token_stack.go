package tokenizer

import "fmt"

type TokenStack[T ~int] struct {
	Levels     []TokenList[T]
	TypeLevels map[T][]int
}

func NewTokenStack[T ~int]() TokenStack[T] {
	return TokenStack[T]{Levels: []TokenList[T]{{}}, TypeLevels: make(map[T][]int)}
}

func (s *TokenStack[T]) Empty() bool {
	return len(s.TypeLevels) == 0 && len(s.Levels) == 1 && len(s.Levels[0]) == 0
}

func (s *TokenStack[T]) LastLevel() *TokenList[T] {
	if len(s.Levels) == 0 {
		return nil
	}
	return &s.Levels[len(s.Levels)-1]
}

func (s *TokenStack[T]) PopCommit() {
	if len(s.Levels) <= 1 {
		panic(fmt.Errorf("unable to pop from the TokenStack with only %v levels", len(s.Levels)))
	}

	popLevel := s.LastLevel()
	s.Levels = s.Levels[0 : len(s.Levels)-1]
	activeLevel := s.LastLevel()

	var firstPosition, lastPosition int
	for i, token := range *popLevel {
		if token.IsDefault() {
			continue
		}
		activeLevel.Push(token)
		if i == 0 {
			firstPosition = len(*activeLevel) - 1
		}
		if i == len(*popLevel)-1 {
			lastPosition = len(*activeLevel) - 1
		}
	}
	if popLevel.FirstOrDefault().Type^Open == popLevel.LastOrDefault().Type {
		(*s.LastLevel())[firstPosition].JumpToPair = lastPosition - firstPosition
		(*s.LastLevel())[lastPosition].JumpToPair = -(lastPosition - firstPosition)
	}
	if typeLevels := s.TypeLevels[popLevel.FirstOrDefault().Type]; len(typeLevels) > 0 {
		s.TypeLevels[popLevel.FirstOrDefault().Type] = typeLevels[0 : len(typeLevels)-1]
	}
}

func (s *TokenStack[T]) PopForget() {
	if len(s.Levels) <= 1 {
		panic(fmt.Errorf("unable to pop from the builder with only %v levels", len(s.Levels)))
	}
	if typeLevels := s.TypeLevels[s.LastLevel().FirstOrDefault().Type]; len(typeLevels) > 0 {
		s.TypeLevels[s.LastLevel().FirstOrDefault().Type] = typeLevels[0 : len(typeLevels)-1]
	}
	popLevel := s.LastLevel()
	s.Levels = s.Levels[0 : len(s.Levels)-1]
	activeLevel := s.LastLevel()
	for _, token := range (*popLevel)[1:] {
		if token.IsDefault() {
			continue
		}
		activeLevel.Push(token)
	}
}

func (s *TokenStack[T]) PopForgetUntil(tokenType T) bool {
	levels := s.TypeLevels[tokenType]
	if len(levels) == 0 {
		return false
	}
	lastLevel := levels[len(levels)-1]
	for len(s.Levels) > lastLevel+1 {
		s.PopForget()
	}
	return true
}

func (s *TokenStack[T]) OpenLevelAt(token Token[T]) {
	s.TypeLevels[token.Type] = append(s.TypeLevels[token.Type], len(s.Levels))
	s.Levels = append(s.Levels, TokenList[T]{token})
}

func (s *TokenStack[T]) CloseLevelAt(token Token[T]) {
	s.LastLevel().Push(token)
	s.PopCommit()
}
