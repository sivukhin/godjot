package djot_parser

import (
	"fmt"
	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/html_writer"
	"github.com/sivukhin/godjot/tokenizer"
)

func ConvertDjotToHtml(builder *html_writer.HtmlWriter, format string, trees ...TreeNode[DjotNode]) {
	for _, tree := range trees {
		switch tree.Type {
		case ThematicBreakNode:
			builder.Tag("hr")
			builder.WriteString("\n")
		case DivNode:
			builder.InTag("div", tree.Attributes.Entries()...)(func() {
				builder.WriteString("\n")
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case UnorderedListNode:
			builder.InTag("ul", tree.Attributes.Entries()...)(func() {
				builder.WriteString("\n")
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case OrderedListNode:
			builder.InTag("ol", tree.Attributes.Entries()...)(func() {
				builder.WriteString("\n")
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case ListItemNode:
			builder.InTag("li", tree.Attributes.Entries()...)(func() {
				builder.WriteString("\n")
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case SectionNode:
			builder.InTag("section", tree.Attributes.Entries()...)(func() {
				builder.WriteString("\n")
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case HeadingNode:
			level := len(tree.Attributes.Get(HeadingLevelKey))
			builder.InTag(fmt.Sprintf("h%v", level), tree.Attributes.Entries()...)(func() {
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case InsertNode:
			builder.InTag("ins", tree.Attributes.Entries()...)(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case SymbolsNode:
			symbol := string(tree.Text)
			switch symbol {
			case "+1":
				builder.WriteString("👍")
			case "smiley":
				builder.WriteString("😃")
			}
		case DeleteNode:
			builder.InTag("del")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case SuperscriptNode:
			builder.InTag("sup")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case SubscriptNode:
			builder.InTag("sub")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case HighlightedNode:
			builder.InTag("mark")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case EmphasisNode:
			builder.InTag("em", tree.Attributes.Entries()...)(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case StrongNode:
			builder.InTag("strong")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case ParagraphNode:
			builder.InTag("p")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
			builder.WriteString("\n")
		case QuoteNode:
			builder.InTag("blockquote")(func() {
				builder.WriteString("\n")
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
			builder.WriteString("\n")
		case CodeNode:
			builder.InTag("pre")(func() {
				builder.InTag("code")(func() {
					ConvertDjotToHtml(builder, format, tree.Children...)
				})
			})
			builder.WriteString("\n")
		case RawNode:
			if format == tree.Attributes.Get(RawBlockFormatKey) {
				ConvertDjotToHtml(builder, format, tree.Children...)
			}
		case DocumentNode:
			ConvertDjotToHtml(builder, format, tree.Children...)
		case ImageNode:
			builder.Tag("img", tree.Attributes.Entries()...)
		case LinkNode:
			builder.InTag("a", tree.Attributes.Entries()...)(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case VerbatimNode:
			if _, ok := tree.Attributes.TryGet(djot_tokenizer.InlineMathKey); ok {
				builder.InTag("span", tokenizer.AttributeEntry{Key: "class", Value: "math inline"})(func() {
					builder.WriteString("\\(")
					builder.WriteBytes(tree.Text)
					builder.WriteString("\\)")
				})
			} else if _, ok := tree.Attributes.TryGet(djot_tokenizer.DisplayMathKey); ok {
				builder.InTag("span", tokenizer.AttributeEntry{Key: "class", Value: "math display"})(func() {
					builder.WriteString("\\[")
					builder.WriteBytes(tree.Text)
					builder.WriteString("\\]")
				})
			} else if rawFormat := tree.Attributes.Get(RawInlineFormatKey); rawFormat == format {
				builder.WriteBytes(tree.Text)
			} else {
				builder.InTag("code")(func() { builder.WriteBytes(tree.Text) })
			}
		case SpanNode:
			builder.InTag("span", tree.Attributes.Entries()...)(func() {
				ConvertDjotToHtml(builder, format, tree.Children...)
			})
		case LineBreakNode:
			builder.Tag("br")
		case TextNode:
			builder.WriteBytes(tree.Text)
		}
	}
}
