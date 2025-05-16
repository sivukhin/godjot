package html_writer

import (
	"strings"

	"maps"

	. "github.com/sivukhin/godjot/djot_parser"
)

type HtmlConversionState = ConversionState[*HtmlWriter]

func StandaloneNodeConverter(state HtmlConversionState, tag string) *HtmlWriter {
	return state.Writer.OpenTag(tag, state.Node.Attributes.Entries()...)
}

func InlineNodeConverter(state HtmlConversionState, tag string, next func(c Children)) *HtmlWriter {
	return state.Writer.InTag(tag, state.Node.Attributes.Entries()...)(func() { next(nil) })
}

func BlockNodeConverter(state HtmlConversionState, tag string, next func(c Children)) *HtmlWriter {
	content := func() {
		state.Writer.WriteString("\n")
		next(nil)
	}
	return state.Writer.InTag(tag, state.Node.Attributes.Entries()...)(content).WriteString("\n")
}

var htmlReplacer = strings.NewReplacer(
	`&`, "&amp;",
	`<`, "&lt;",
	`>`, "&gt;",
	`–`, `&ndash;`,
	`—`, `&mdash;`,
	`“`, `&ldquo;`,
	`”`, `&rdquo;`,
	`‘`, `&lsquo;`,
	`’`, `&rsquo;`,
	`…`, `&hellip;`,
)

func NewHtmlConversionContext(format string, converters ...map[DjotNode]Conversion[*HtmlWriter]) ConversionContext[*HtmlWriter] {
	if len(converters) == 0 {
		converters = []map[DjotNode]Conversion[*HtmlWriter]{HtmlConversionRegistry}
	}
	registry := make(map[DjotNode]Conversion[*HtmlWriter])
	for i := range converters {
		maps.Copy(registry, converters[i])
	}
	return ConversionContext[*HtmlWriter]{
		Format:   format,
		Registry: registry,
	}
}

func ConvertDjotToHtml(
	context ConversionContext[*HtmlWriter],
	builder *HtmlWriter,
	nodes ...TreeNode[DjotNode],
) string {
	convertDjotToHtml(context, builder, nil, nodes...)
	return builder.String()
}

func convertDjotToHtml(context ConversionContext[*HtmlWriter], builder *HtmlWriter, parent *TreeNode[DjotNode], nodes ...TreeNode[DjotNode]) {
	for _, node := range nodes {
		currentNode := node
		conversion, ok := context.Registry[currentNode.Type]
		if !ok {
			continue
		}
		state := ConversionState[*HtmlWriter]{
			Format: context.Format,
			Writer: builder,
			Node:   currentNode,
			Parent: parent,
		}
		conversion(state, func(c Children) {
			if len(c) == 0 {
				convertDjotToHtml(context, builder, &node, currentNode.Children...)
			} else {
				convertDjotToHtml(context, builder, &node, c...)
			}
		})
	}
}
