package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenStack(t *testing.T) {
	t.Run("open 2, push 10, close 2", func(t *testing.T) {
		s := NewTokenStack[int]()
		s.OpenLevelAt(Token[int]{Type: 2, Start: 0, End: 1})
		s.LastLevel().Push(Token[int]{Type: 10, Start: 1, End: 4})
		s.CloseLevelAt(Token[int]{Type: 2 ^ Open, Start: 10, End: 11})
		require.Equal(t, &TokenList[int]{
			{Type: 2, Start: 0, End: 1, JumpToPair: 3},
			{Type: 10, Start: 1, End: 4},
			{Type: 0, Start: 4, End: 10},
			{Type: 2 ^ Open, Start: 10, End: 11, JumpToPair: -3},
		}, s.LastLevel())
	})
	t.Run("open 2, open 4, open 6, close 4, pop", func(t *testing.T) {
		s := NewTokenStack[int]()
		s.OpenLevelAt(Token[int]{Type: 2, Start: 0, End: 0})
		s.OpenLevelAt(Token[int]{Type: 4, Start: 0, End: 1})
		s.OpenLevelAt(Token[int]{Type: 6, Start: 1, End: 2})
		s.OpenLevelAt(Token[int]{Type: 8, Start: 2, End: 3})
		require.False(t, s.PopForgetUntil(10))
		require.True(t, s.PopForgetUntil(6))
		s.CloseLevelAt(Token[int]{Type: 6 ^ Open, Start: 10, End: 11})
		s.PopForget()
		s.CloseLevelAt(Token[int]{Type: 2 ^ Open, Start: 11, End: 11})
		require.Equal(t, &TokenList[int]{
			{Type: 2, Start: 0, End: 0, JumpToPair: 5},
			{Type: 0, Start: 0, End: 1},
			{Type: 6, Start: 1, End: 2, JumpToPair: 2},
			{Type: 0, Start: 2, End: 10},
			{Type: 6 ^ Open, Start: 10, End: 11, JumpToPair: -2},
			{Type: 2 ^ Open, Start: 11, End: 11, JumpToPair: -5},
		}, s.LastLevel())
	})
}
