package djot_parser

import (
	"bytes"
	"fmt"
	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/html_writer"
	"github.com/sivukhin/godjot/tokenizer"
)

func ConvertDjotToHtml(builder *html_writer.HtmlWriter, format string, trees ...Tree[DjotNode]) {
	for _, tree := range trees {
		switch tree.Type {
		case HeadingNode:
			level := len(bytes.TrimSuffix(tree.Token, []byte(" ")))
			builder.InTag(fmt.Sprintf("h%v", level))(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
		case InsertNode:
			builder.InTag("ins")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
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
			builder.InTag("em")(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
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
		case DocumentNode:
			ConvertDjotToHtml(builder, format, tree.Children...)
		case ImageNode:
			attributes := tree.Attributes.Entries()
			if bytes.Contains(tree.Text, []byte("@")) {
				attributes = append(attributes, tokenizer.AttributeEntry{Key: "src", Value: "mailto:" + string(tree.Text)})
			} else if len(tree.Text) > 0 {
				attributes = append(attributes, tokenizer.AttributeEntry{Key: "src", Value: string(tree.Text)})
			}
			builder.Tag("img", attributes...)
		case LinkNode:
			var attributes []tokenizer.AttributeEntry
			if bytes.Contains(tree.Text, []byte("@")) {
				attributes = append(attributes, tokenizer.AttributeEntry{Key: "href", Value: "mailto:" + string(tree.Text)})
			} else if len(tree.Text) > 0 {
				attributes = append(attributes, tokenizer.AttributeEntry{Key: "href", Value: string(tree.Text)})
			}
			builder.InTag("a", attributes...)(func() { ConvertDjotToHtml(builder, format, tree.Children...) })
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
			} else if rawFormat := tree.Attributes.Get(RawFormatKey); rawFormat == format {
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
