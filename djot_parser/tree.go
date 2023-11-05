package djot_parser

import "github.com/sivukhin/godjot/tokenizer"

type Tree[T ~int] struct {
	Type        T
	Attributes  *tokenizer.Attributes
	Children    []Tree[T]
	Token, Text []byte
}
