//go:build !debug_on && !test

package tokenizer

func Assertf(condition bool, format string, args ...any) {}
