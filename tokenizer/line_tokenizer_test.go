package tokenizer

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLineTokenizer(t *testing.T) {
	document := []byte("hello\nworld\n!")
	tokenizer := LineTokenizer{Document: document}
	{
		start, end, eof := tokenizer.Scan()
		require.False(t, eof)
		require.Equal(t, "hello\n", string(document[start:end]))
	}
	{
		start, end, eof := tokenizer.Scan()
		require.False(t, eof)
		require.Equal(t, "world\n", string(document[start:end]))
	}
	{
		start, end, eof := tokenizer.Scan()
		require.False(t, eof)
		require.Equal(t, "!", string(document[start:end]))
	}
	{
		_, _, eof := tokenizer.Scan()
		require.True(t, eof)
	}
}
