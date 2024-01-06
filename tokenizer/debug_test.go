package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssert(t *testing.T) {
	require.Panics(t, func() { Assertf(false, "expected true") })
}
