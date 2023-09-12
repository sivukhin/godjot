package parser

type Tree[T ~int] struct {
	Type       T
	Attributes map[string]string
	Children   []Tree[T]
	Text       string
}
