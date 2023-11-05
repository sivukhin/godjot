package html_writer

import (
	"github.com/sivukhin/godjot/tokenizer"
	"sort"
	"strings"
)

type HtmlWriter struct {
	Builder strings.Builder
}

func (w *HtmlWriter) String() string { return w.Builder.String() }

func (w *HtmlWriter) Tag(tag string, attributes ...tokenizer.AttributeEntry) {
	w.Builder.WriteString("<")
	w.Builder.WriteString(tag)
	for _, attribute := range attributes {
		w.Builder.WriteString(" ")
		w.Builder.WriteString(attribute.Key)
		w.Builder.WriteString("=\"")
		w.Builder.WriteString(attribute.Value)
		w.Builder.WriteString("\"")
	}
	w.Builder.WriteString(">")
}

func (w *HtmlWriter) InTag(tag string, attributes ...tokenizer.AttributeEntry) func(func()) {
	return func(content func()) {
		w.Builder.WriteString("<")
		w.Builder.WriteString(tag)
		sort.Slice(attributes, func(i, j int) bool {
			iStart := attributes[i].Key
			jStart := attributes[j].Key
			if iStart == "id" && jStart != "id" {
				return true
			}
			if iStart != "id" && jStart == "id" {
				return false
			}
			if iStart == "class" && jStart != "class" {
				return true
			}
			if iStart != "class" && jStart == "class" {
				return false
			}
			return i < j
		})
		for _, attribute := range attributes {
			w.Builder.WriteString(" ")
			w.Builder.WriteString(attribute.Key)
			w.Builder.WriteString("=\"")
			w.Builder.WriteString(attribute.Value)
			w.Builder.WriteString("\"")
		}
		w.Builder.WriteString(">")
		content()
		w.Builder.WriteString("</")
		w.Builder.WriteString(tag)
		w.Builder.WriteString(">")
	}
}

func (w *HtmlWriter) WriteBytes(text []byte)  { w.Builder.Write(text) }
func (w *HtmlWriter) WriteString(text string) { w.Builder.WriteString(text) }
