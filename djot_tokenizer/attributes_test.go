package djot_tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"git.sr.ht/~ser/godjot/v2/tokenizer"
)

func TestQuotedString(t *testing.T) {
	t.Run("matched", func(t *testing.T) {
		for _, tt := range []struct {
			s     string
			value []byte
		}{
			{s: `"hello"`, value: []byte(`hello`)},
			{s: `""`, value: nil},
			{s: `"this is (\") quote"`, value: []byte(`this is (") quote`)},
		} {
			t.Run(tt.s, func(t *testing.T) {
				reader := tokenizer.TextReader(tt.s)
				value, next, ok := MatchQuotedString(reader, 0)
				require.True(t, ok)
				require.Equal(t, len(tt.s), int(next))
				require.Equal(t, tt.value, value)
			})
		}
	})
	t.Run("unmatched", func(t *testing.T) {
		for _, tt := range []string{`"hello`, `"hello\"`, `hello`, "`hello`"} {
			t.Run(tt, func(t *testing.T) {
				reader := tokenizer.TextReader(tt)
				_, _, ok := MatchQuotedString(reader, 0)
				require.False(t, ok)
			})
		}
	})
}

func TestAttributes(t *testing.T) {
	t.Run("matched", func(t *testing.T) {
		for _, tt := range []struct {
			s     string
			value map[string]string
		}{
			{s: `{% This is a comment, spanning\nmultiple lines %}`, value: nil},
			{s: `{.some-class}`, value: map[string]string{DjotAttributeClassKey: "some-class"}},
			{s: `{.some-class % comment \n with \n newlines %}`, value: map[string]string{DjotAttributeClassKey: "some-class"}},
			{s: `{.a % comment % .b}`, value: map[string]string{DjotAttributeClassKey: "a b"}},
			{s: `{#some-id}`, value: map[string]string{DjotAttributeIdKey: "some-id"}},
			{s: `{some-key=some-value}`, value: map[string]string{"some-key": "some-value"}},
			{s: `{some-key="left \"middle\" right"}`, value: map[string]string{"some-key": "left \"middle\" right"}},
			{s: `{ .a    .b   }`, value: map[string]string{DjotAttributeClassKey: "a b"}}} {
			t.Run(tt.s, func(t *testing.T) {
				reader := tokenizer.TextReader(tt.s)
				value, next, ok := MatchDjotAttribute(reader, 0)
				require.True(t, ok)
				require.Equal(t, len(tt.s), int(next))
				require.Equal(t, tt.value, value.GoMap())
			})
		}
	})
}
