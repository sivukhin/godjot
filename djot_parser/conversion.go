package djot_parser

type (
	// ConversionContext holds registry of conversion functions for all AST nodes
	// Note, that it has generic parameter T which is opaque for the library and caller can use it however he wants
	// (for example, render data somewhere or just analyze AST and accumulate some knowledge in the internal fields of T)
	//
	// Also, third-party libraries can implement custom factories for ConversionContext for targets different from HTML
	// (see https://github.com/sivukhin/godjot/issues/14 for more details)
	ConversionContext[T any] struct {
		Format   string
		Registry ConversionRegistry[T]
	}
	ConversionState[T any] struct {
		Format string
		Writer T
		Node   TreeNode[DjotNode]
		Parent *TreeNode[DjotNode]
	}
	Conversion[T any]         func(state ConversionState[T], next func(Children))
	ConversionRegistry[T any] map[DjotNode]Conversion[T]
	Children                  []TreeNode[DjotNode]
)

func (context ConversionContext[T]) ConvertDjot(
	builder T,
	nodes ...TreeNode[DjotNode],
) T {
	context.convertDjot(builder, nil, nodes...)
	return builder
}

func (context ConversionContext[T]) convertDjot(
	builder T,
	parent *TreeNode[DjotNode],
	nodes ...TreeNode[DjotNode],
) {
	for _, node := range nodes {
		currentNode := node
		conversion, ok := context.Registry[currentNode.Type]
		if !ok {
			continue
		}
		state := ConversionState[T]{
			Format: context.Format,
			Writer: builder,
			Node:   currentNode,
			Parent: parent,
		}
		conversion(state, func(c Children) {
			if len(c) == 0 {
				context.convertDjot(builder, &node, currentNode.Children...)
			} else {
				context.convertDjot(builder, &node, c...)
			}
		})
	}
}
