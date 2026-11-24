package djot_html

import (
	"strings"

	"maps"

	. "git.sr.ht/~ser/godjot/v2/djot_parser"
)

func StandaloneNodeConverter(state ConversionState[*HtmlWriter], tag string) *HtmlWriter {
	return state.Writer.OpenTag(tag, state.Node.Attributes.Entries()...)
}

func InlineNodeConverter(state ConversionState[*HtmlWriter], tag string, next func(c Children)) *HtmlWriter {
	return state.Writer.InTag(tag, state.Node.Attributes.Entries()...)(func() { next(nil) })
}

func BlockNodeConverter(state ConversionState[*HtmlWriter], tag string, next func(c Children)) *HtmlWriter {
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

func New(converters ...map[DjotNode]Conversion[*HtmlWriter]) ConversionContext[*HtmlWriter] {
	if len(converters) == 0 {
		converters = []map[DjotNode]Conversion[*HtmlWriter]{DefaultConversionRegistry}
	}
	registry := make(map[DjotNode]Conversion[*HtmlWriter])
	for i := range converters {
		maps.Copy(registry, converters[i])
	}
	return ConversionContext[*HtmlWriter]{
		Format:   "html",
		Registry: registry,
	}
}
