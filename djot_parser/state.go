package djot_parser

import (
	"io"
)

type (
	ConversionContext[T io.Writer] struct {
		Format   string
		Registry ConversionRegistry[T]
	}
	ConversionState[T io.Writer] struct {
		Format string
		Writer T
		Node   TreeNode[DjotNode]
		Parent *TreeNode[DjotNode]
	}
	Conversion[T io.Writer]         func(state ConversionState[T], next func(Children))
	ConversionRegistry[T io.Writer] map[DjotNode]Conversion[T]
	Children                        []TreeNode[DjotNode]
)

var DefaultSymbolRegistry = map[string]string{}

// NewConversionContext is made publicly available in order to allow third-party libraries to implement custom render of Djot markdown to the target different from HTML
// (see https://github.com/sivukhin/godjot/issues/14 for more details)
func NewConversionContext[T io.Writer](format string, converters ...map[DjotNode]Conversion[T]) ConversionContext[T] {
	if len(converters) == 0 {
		converters = []map[DjotNode]Conversion[T]{{
			HeadingNode:        func(state ConversionState[T], next func(c Children)) {},
			SectionNode:        func(state ConversionState[T], next func(c Children)) {},
			TaskListNode:       func(state ConversionState[T], next func(c Children)) {},
			DefinitionTermNode: func(state ConversionState[T], next func(c Children)) {},
			DefinitionItemNode: func(state ConversionState[T], next func(c Children)) {},
			RawNode:            func(state ConversionState[T], next func(c Children)) {},
			ThematicBreakNode:  func(state ConversionState[T], next func(c Children)) {},
			DivNode:            func(state ConversionState[T], next func(c Children)) {},
			TableCaptionNode:   func(state ConversionState[T], next func(c Children)) {},
			ReferenceDefNode:   func(state ConversionState[T], next func(c Children)) {},
			FootnoteDefNode:    func(state ConversionState[T], next func(c Children)) {},
			HighlightedNode:    func(state ConversionState[T], next func(c Children)) {},
			SubscriptNode:      func(state ConversionState[T], next func(c Children)) {},
			SuperscriptNode:    func(state ConversionState[T], next func(c Children)) {},
			InsertNode:         func(state ConversionState[T], next func(c Children)) {},
			SymbolsNode:        func(state ConversionState[T], next func(c Children)) {},
			VerbatimNode:       func(state ConversionState[T], next func(c Children)) {},
			LineBreakNode:      func(state ConversionState[T], next func(c Children)) {},
			SpanNode:           func(state ConversionState[T], next func(c Children)) {},
			DocumentNode:       func(state ConversionState[T], next func(c Children)) {},
			QuoteNode:          func(state ConversionState[T], next func(c Children)) {},
			OrderedListNode:    func(state ConversionState[T], next func(c Children)) {},
			UnorderedListNode:  func(state ConversionState[T], next func(c Children)) {},
			DefinitionListNode: func(state ConversionState[T], next func(c Children)) {},
			ListItemNode:       func(state ConversionState[T], next func(c Children)) {},
			ParagraphNode:      func(state ConversionState[T], next func(c Children)) {},
			EmphasisNode:       func(state ConversionState[T], next func(c Children)) {},
			StrongNode:         func(state ConversionState[T], next func(c Children)) {},
			DeleteNode:         func(state ConversionState[T], next func(c Children)) {},
			LinkNode:           func(state ConversionState[T], next func(c Children)) {},
			ImageNode:          func(state ConversionState[T], next func(c Children)) {},
			TextNode:           func(state ConversionState[T], next func(c Children)) {},
			CodeNode:           func(state ConversionState[T], next func(c Children)) {},
			TableNode:          func(state ConversionState[T], next func(c Children)) {},
			TableCellNode:      func(state ConversionState[T], next func(c Children)) {},
			TableHeaderNode:    func(state ConversionState[T], next func(c Children)) {},
			TableRowNode:       func(state ConversionState[T], next func(c Children)) {},
		}}
	}
	registry := make(map[DjotNode]Conversion[T])
	for i := 0; i < len(converters); i++ {
		for node, conversion := range converters[i] {
			registry[node] = conversion
		}
	}
	return ConversionContext[T]{
		Format:   format,
		Registry: registry,
	}
}
