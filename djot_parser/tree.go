package djot_parser

import "github.com/sivukhin/godjot/tokenizer"

type TreeNode[T ~int] struct {
	Type       T
	Attributes *tokenizer.Attributes
	Children   []TreeNode[T]
	Text       []byte
}
