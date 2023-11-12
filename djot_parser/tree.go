package djot_parser

import "github.com/sivukhin/godjot/tokenizer"

type TreeNode[T ~int] struct {
	Type       T
	Attributes *tokenizer.Attributes
	Children   []TreeNode[T]
	Text       []byte
}

func (n TreeNode[T]) Traverse(f func(node TreeNode[T])) {
	f(n)
	for _, child := range n.Children {
		child.Traverse(f)
	}
}
