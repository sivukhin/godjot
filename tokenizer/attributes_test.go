package tokenizer

import (
	"github.com/stretchr/testify/require"
	"testing"
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
				reader := TextReader(tt.s)
				value, next := MatchQuotedString(reader, 0)
				require.Equal(t, len(tt.s), int(next))
				require.Equal(t, tt.value, value)
			})
		}
	})
	t.Run("unmatched", func(t *testing.T) {
		for _, tt := range []string{`"hello`, `"hello\"`, `hello`, "`hello`"} {
			t.Run(tt, func(t *testing.T) {
				reader := TextReader(tt)
				_, next := MatchQuotedString(reader, 0)
				require.Equal(t, Unmatched, next)
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
			{s: `{.some-class}`, value: map[string]string{DjotAttributeClassKey: "some-class"}},
			{s: `{.a % comment % .b}`, value: map[string]string{DjotAttributeClassKey: "a b"}},
			{s: `{#some-id}`, value: map[string]string{DjotAttributeIdKey: "some-id"}},
			{s: `{some-key=some-value}`, value: map[string]string{"some-key": "some-value"}},
			{s: `{some-key="left \"middle\" right"}`, value: map[string]string{"some-key": "left \"middle\" right"}},
			{s: `{ .a    .b   }`, value: map[string]string{DjotAttributeClassKey: "a b"}}} {
			t.Run(tt.s, func(t *testing.T) {
				reader := TextReader(tt.s)
				value, next := MatchDjotAttribute(reader, 0)
				require.Equal(t, len(tt.s), int(next))
				require.Equal(t, tt.value, value)
			})
		}
	})
}
