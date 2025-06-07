package djot_parser

import "github.com/sivukhin/godjot/v2/tokenizer"

type TreeNode[T ~int] struct {
	Type       T
	Attributes tokenizer.Attributes
	Children   []TreeNode[T]
	Text       []byte
	Index      int
}

func (n TreeNode[T]) Traverse(f func(node TreeNode[T])) {
	f(n)
	for _, child := range n.Children {
		child.Traverse(f)
	}
}

func (n TreeNode[T]) FullText() []byte {
	textNodes := 0
	var text []byte
	n.Traverse(func(node TreeNode[T]) {
		textNodes += 1
		if textNodes == 1 {
			text = node.Text // optimization to avoid unnecessary allocations
			return
		}
		if textNodes == 2 {
			text = append([]byte{}, text...)
		}
		text = append(text, node.Text...)
	})
	return text
}
