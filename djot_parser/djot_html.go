package djot_parser

import (
	"fmt"

	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/html_writer"
	"github.com/sivukhin/godjot/tokenizer"
)

type (
	ConversionContext struct {
		Format   string
		Registry ConversionRegistry
	}
	ConversionState struct {
		Format string
		Writer *html_writer.HtmlWriter
		Node   TreeNode[DjotNode]
	}
	Conversion         func(state ConversionState, next func())
	ConversionRegistry map[DjotNode]Conversion
)

func (state ConversionState) StandaloneNodeConverter(tag string) *html_writer.HtmlWriter {
	return state.Writer.OpenTag(tag, state.Node.Attributes.Entries()...)
}

func (state ConversionState) InlineNodeConverter(tag string, next func()) *html_writer.HtmlWriter {
	return state.Writer.InTag(tag, state.Node.Attributes.Entries()...)(next)
}

func (state ConversionState) BlockNodeConverter(tag string, next func()) *html_writer.HtmlWriter {
	content := func() {
		state.Writer.WriteString("\n")
		next()
	}
	return state.Writer.InTag(tag, state.Node.Attributes.Entries()...)(content).WriteString("\n")
}

var DefaultSymbolRegistry = map[string]string{
	"+1":     "👍",
	"smiley": "😃",
}

var DefaultConversionRegistry = map[DjotNode]Conversion{
	ThematicBreakNode: func(s ConversionState, n func()) { s.Writer.OpenTag("th").WriteString("\n") },
	LineBreakNode:     func(s ConversionState, n func()) { s.Writer.OpenTag("br").WriteString("\n") },
	TextNode:          func(s ConversionState, n func()) { s.Writer.WriteBytes(s.Node.Text) },
	SymbolsNode:       func(s ConversionState, n func()) { s.Writer.WriteString(DefaultSymbolRegistry[string(s.Node.Text)]) },
	InsertNode:        func(s ConversionState, n func()) { s.InlineNodeConverter("ins", n) },
	DeleteNode:        func(s ConversionState, n func()) { s.InlineNodeConverter("del", n) },
	SuperscriptNode:   func(s ConversionState, n func()) { s.InlineNodeConverter("sup", n) },
	SubscriptNode:     func(s ConversionState, n func()) { s.InlineNodeConverter("sub", n) },
	HighlightedNode:   func(s ConversionState, n func()) { s.InlineNodeConverter("mark", n) },
	EmphasisNode:      func(s ConversionState, n func()) { s.InlineNodeConverter("em", n) },
	StrongNode:        func(s ConversionState, n func()) { s.InlineNodeConverter("strong", n) },
	ParagraphNode:     func(s ConversionState, n func()) { s.InlineNodeConverter("p", n).WriteString("\n") },
	ImageNode:         func(s ConversionState, n func()) { s.StandaloneNodeConverter("img") },
	LinkNode:          func(s ConversionState, n func()) { s.InlineNodeConverter("a", n) },
	SpanNode:          func(s ConversionState, n func()) { s.InlineNodeConverter("span", n) },
	DivNode:           func(s ConversionState, n func()) { s.BlockNodeConverter("div", n) },
	UnorderedListNode: func(s ConversionState, n func()) { s.BlockNodeConverter("ul", n) },
	OrderedListNode:   func(s ConversionState, n func()) { s.BlockNodeConverter("ol", n) },
	ListItemNode:      func(s ConversionState, n func()) { s.BlockNodeConverter("li", n) },
	SectionNode:       func(s ConversionState, n func()) { s.BlockNodeConverter("section", n) },
	QuoteNode:         func(s ConversionState, n func()) { s.BlockNodeConverter("blockquote", n) },
	DocumentNode:      func(s ConversionState, n func()) { n() },
	CodeNode: func(s ConversionState, n func()) {
		s.Writer.OpenTag("pre", s.Node.Attributes.Entries()...).OpenTag("code")
		n()
		s.Writer.CloseTag("code").CloseTag("pre").WriteString("\n")
	},
	VerbatimNode: func(s ConversionState, n func()) {
		if _, ok := s.Node.Attributes.TryGet(djot_tokenizer.InlineMathKey); ok {
			s.Writer.InTag("span", tokenizer.AttributeEntry{Key: "class", Value: "math inline"})(func() {
				s.Writer.WriteString("\\(")
				s.Writer.WriteBytes(s.Node.Text)
				s.Writer.WriteString("\\)")
			})
		} else if _, ok := s.Node.Attributes.TryGet(djot_tokenizer.DisplayMathKey); ok {
			s.Writer.InTag("span", tokenizer.AttributeEntry{Key: "class", Value: "math display"})(func() {
				s.Writer.WriteString("\\[")
				s.Writer.WriteBytes(s.Node.Text)
				s.Writer.WriteString("\\]")
			})
		} else if rawFormat := s.Node.Attributes.Get(RawInlineFormatKey); rawFormat == s.Format {
			s.Writer.WriteBytes(s.Node.Text)
		} else {
			s.Writer.InTag("code")(func() { s.Writer.WriteBytes(s.Node.Text) })
		}
	},
	HeadingNode: func(s ConversionState, n func()) {
		level := len(s.Node.Attributes.Get(HeadingLevelKey))
		s.Writer.InTag(fmt.Sprintf("h%v", level), s.Node.Attributes.Entries()...)(func() { n() }).WriteString("\n")
	},
	RawNode: func(s ConversionState, next func()) {
		if s.Format == s.Node.Attributes.Get(RawBlockFormatKey) {
			next()
		}
	},
}

func NewConversionContext(format string, converters ...map[DjotNode]Conversion) ConversionContext {
	if len(converters) == 0 {
		converters = []map[DjotNode]Conversion{DefaultConversionRegistry}
	}
	registry := converters[0]
	for i := 1; i < len(converters); i++ {
		for node, conversion := range converters[i] {
			registry[node] = conversion
		}
	}
	return ConversionContext{
		Format:   format,
		Registry: registry,
	}
}

func (context ConversionContext) ConvertDjotToHtml(nodes ...TreeNode[DjotNode]) string {
	builder := html_writer.HtmlWriter{}
	context.convertDjotToHtml(&builder, nodes...)
	return builder.String()
}

func (context ConversionContext) convertDjotToHtml(builder *html_writer.HtmlWriter, nodes ...TreeNode[DjotNode]) {
	for _, node := range nodes {
		currentNode := node
		conversion, ok := context.Registry[currentNode.Type]
		if !ok {
			continue
		}
		state := ConversionState{
			Format: context.Format,
			Writer: builder,
			Node:   currentNode,
		}
		conversion(state, func() { context.convertDjotToHtml(builder, currentNode.Children...) })
	}
}
