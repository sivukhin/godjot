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

func (n TreeNode[T]) FullText() []byte {
	var text []byte
	n.Traverse(func(node TreeNode[T]) {
		if len(text) == 0 {
			text = node.Text
		} else if len(node.Text) > 0 {
			text = append(text, node.Text...)
		}
	})
	return text
}
