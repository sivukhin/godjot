package html_writer

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	. "github.com/sivukhin/godjot/djot_parser"
	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/tokenizer"
)

var HtmlConversionRegistry = map[DjotNode]Conversion[*HtmlWriter]{
	ThematicBreakNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { s.Writer.OpenTag("hr").WriteString("\n") },
	LineBreakNode:     func(s ConversionState[*HtmlWriter], n func(c Children)) { s.Writer.OpenTag("br").WriteString("\n") },
	TextNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		if s.Parent != nil && (s.Parent.Attributes.Get(RawInlineFormatKey) == s.Format || s.Parent.Attributes.Get(RawBlockFormatKey) == s.Format) {
			s.Writer.WriteString(string(s.Node.Text))
		} else {
			s.Writer.WriteString(htmlReplacer.Replace(string(s.Node.Text)))
		}
	},
	SymbolsNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		symbol, ok := DefaultSymbolRegistry[string(s.Node.FullText())]
		if ok {
			s.Writer.WriteString(symbol)
		} else {
			s.Writer.WriteString(fmt.Sprintf(":%v:", string(s.Node.FullText())))
		}
	},
	InsertNode:      func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "ins", n) },
	DeleteNode:      func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "del", n) },
	SuperscriptNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "sup", n) },
	SubscriptNode:   func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "sub", n) },
	HighlightedNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "mark", n) },
	EmphasisNode:    func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "em", n) },
	StrongNode:      func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "strong", n) },
	ParagraphNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		InlineNodeConverter(s, "p", n).WriteString("\n")
	},
	ImageNode:        func(s ConversionState[*HtmlWriter], n func(c Children)) { StandaloneNodeConverter(s, "img") },
	LinkNode:         func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "a", n) },
	SpanNode:         func(s ConversionState[*HtmlWriter], n func(c Children)) { InlineNodeConverter(s, "span", n) },
	DivNode:          func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "div", n) },
	TableCaptionNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { n(nil) },
	TableNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		if len(s.Node.Children) > 0 && s.Node.Children[0].Type == TableCaptionNode {
			s.Writer.OpenTag("table", s.Node.Attributes.Entries()...)
			s.Writer.WriteString("\n")
			s.Writer.InTag("caption")(func() { n(s.Node.Children[:1]) })
			s.Writer.WriteString("\n")
			s.Writer.InTag("tbody")(func() { n(s.Node.Children[1:]) })
			s.Writer.CloseTag("table")
		} else {
			BlockNodeConverter(s, "table", n)
		}
	},
	TableRowNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "tr", n) },
	TableHeaderNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		InlineNodeConverter(s, "th", n)
		s.Writer.WriteString("\n")
	},
	TableCellNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		InlineNodeConverter(s, "td", n)
		s.Writer.WriteString("\n")
	},
	TaskListNode:       func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "ul", n) },
	DefinitionListNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "dl", n) },
	UnorderedListNode:  func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "ul", n) },
	OrderedListNode:    func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "ol", n) },
	ListItemNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		class := s.Node.Attributes.Get(djot_tokenizer.DjotAttributeClassKey)
		if class == CheckedTaskItemClass || class == UncheckedTaskItemClass {
			s.Writer.InTag("li")(func() {
				s.Writer.WriteString("\n")
				s.Writer.WriteString("<input disabled=\"\" type=\"checkbox\"")
				if class == CheckedTaskItemClass {
					s.Writer.WriteString(" checked=\"\"")
				}
				s.Writer.WriteString("/>").WriteString("\n")
				n(s.Node.Children)
			}).WriteString("\n")
		} else {
			BlockNodeConverter(s, "li", n)
		}
	},
	DefinitionTermNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		InlineNodeConverter(s, "dt", n)
		s.Writer.WriteString("\n")
	},
	DefinitionItemNode: func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "dd", n) },
	SectionNode:        func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "section", n) },
	QuoteNode:          func(s ConversionState[*HtmlWriter], n func(c Children)) { BlockNodeConverter(s, "blockquote", n) },
	DocumentNode:       func(s ConversionState[*HtmlWriter], n func(c Children)) { n(nil) },
	FootnoteDefNode:    func(s ConversionState[*HtmlWriter], n func(c Children)) { n(nil) },
	CodeNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		s.Writer.OpenTag("pre").OpenTag("code", s.Node.Attributes.Entries()...)
		n(nil)
		s.Writer.CloseTag("code").CloseTag("pre").WriteString("\n")
	},
	VerbatimNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		if _, ok := s.Node.Attributes.TryGet(djot_tokenizer.InlineMathKey); ok {
			attributes := append([]tokenizer.AttributeEntry{{Key: "class", Value: "math inline"}}, s.Node.Attributes.Entries()...)
			s.Writer.InTag("span", attributes...)(func() {
				s.Writer.WriteString("\\(")
				n(nil)
				s.Writer.WriteString("\\)")
			})
		} else if _, ok := s.Node.Attributes.TryGet(djot_tokenizer.DisplayMathKey); ok {
			attributes := append([]tokenizer.AttributeEntry{{Key: "class", Value: "math display"}}, s.Node.Attributes.Entries()...)
			s.Writer.InTag("span", attributes...)(func() {
				s.Writer.WriteString("\\[")
				n(nil)
				s.Writer.WriteString("\\]")
			})
		} else if rawFormat := s.Node.Attributes.Get(RawInlineFormatKey); rawFormat == s.Format {
			n(nil)
		} else {
			s.Writer.InTag("code", s.Node.Attributes.Entries()...)(func() { n(nil) })
		}
	},
	HeadingNode: func(s ConversionState[*HtmlWriter], n func(c Children)) {
		level := len(s.Node.Attributes.Get(HeadingLevelKey))
		s.Writer.InTag(fmt.Sprintf("h%v", level), s.Node.Attributes.Entries()...)(func() { n(nil) }).WriteString("\n")
	},
	RawNode: func(s ConversionState[*HtmlWriter], next func(c Children)) {
		if s.Format == s.Node.Attributes.Get(RawBlockFormatKey) {
			next(nil)
		}
	},
}

type HtmlWriter struct {
	Builder     strings.Builder
	Indentation int
	TabSize     int
	InContent   bool
	InPre       bool
}

func (w *HtmlWriter) String() string { return w.Builder.String() }

func (w *HtmlWriter) OpenTag(tag string, attributes ...tokenizer.AttributeEntry) *HtmlWriter {
	if !w.InContent && !w.InPre {
		w.WriteString(ident(w.Indentation))
	}
	w.Builder.WriteString("<")
	w.Builder.WriteString(tag)
	sort.Slice(attributes, func(i, j int) bool {
		iStart := attributes[i].Key
		jStart := attributes[j].Key
		if iStart == "class" && jStart != "class" {
			return true
		}
		if iStart != "class" && jStart == "class" {
			return false
		}
		if iStart == "id" && jStart != "id" {
			return true
		}
		if iStart != "id" && jStart == "id" {
			return false
		}
		return i < j
	})
	for _, attribute := range attributes {
		if strings.HasPrefix(attribute.Key, "$") {
			continue
		}
		w.Builder.WriteString(" ")
		w.Builder.WriteString(attribute.Key)
		w.Builder.WriteString("=\"")
		w.Builder.WriteString(attribute.Value)
		w.Builder.WriteString("\"")
	}
	w.Builder.WriteString(">")
	w.Indentation += w.TabSize
	w.InContent = true
	if tag == "pre" {
		w.InPre = true
	}
	return w
}

func (w *HtmlWriter) CloseTag(tag string) *HtmlWriter {
	w.Indentation -= w.TabSize
	if !w.InContent && !w.InPre {
		w.WriteString(ident(w.Indentation))
	}
	w.Builder.WriteString("</")
	w.Builder.WriteString(tag)
	w.Builder.WriteString(">")
	if tag == "pre" {
		w.InPre = false
	}
	return w
}

func (w *HtmlWriter) InTag(tag string, attributes ...tokenizer.AttributeEntry) func(func()) *HtmlWriter {
	return func(content func()) *HtmlWriter {
		w.OpenTag(tag, attributes...)
		content()
		w.CloseTag(tag)
		return w
	}
}

func (w *HtmlWriter) WriteBytes(text []byte) *HtmlWriter {
	w.Builder.Write(text)
	w.InContent = !bytes.Equal(text, []byte("\n"))
	return w
}

func (w *HtmlWriter) WriteString(text string) *HtmlWriter {
	w.Builder.WriteString(text)
	w.InContent = text != "\n"
	return w
}

func ident(n int) string {
	return strings.Repeat(" ", n)
}
