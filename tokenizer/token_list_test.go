package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenList(t *testing.T) {
	l := make(TokenList[int], 0)
	l.Push(Token[int]{Type: 1, Start: 0, End: 1})
	l.Push(Token[int]{Type: 2, Start: 10, End: 11})
	require.Equal(t, TokenList[int]{
		{Type: 1, Start: 0, End: 1},
		{Type: 0, Start: 1, End: 10},
		{Type: 2, Start: 10, End: 11},
	}, l)
}
