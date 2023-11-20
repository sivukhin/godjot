package html_writer

import (
	"sort"
	"strings"

	"github.com/sivukhin/godjot/tokenizer"
)

type HtmlWriter struct {
	Builder strings.Builder
}

func (w *HtmlWriter) String() string { return w.Builder.String() }

func (w *HtmlWriter) OpenTag(tag string, attributes ...tokenizer.AttributeEntry) *HtmlWriter {
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
	return w
}

func (w *HtmlWriter) CloseTag(tag string) *HtmlWriter {
	w.Builder.WriteString("</")
	w.Builder.WriteString(tag)
	w.Builder.WriteString(">")
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
	return w
}

func (w *HtmlWriter) WriteString(text string) *HtmlWriter {
	w.Builder.WriteString(text)
	return w
}
