package djot_parser

type Tree[T ~int] struct {
	Type        T
	Attributes  map[string]string
	Children    []Tree[T]
	Token, Text []byte
}
