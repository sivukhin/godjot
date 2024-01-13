package html_writer

import (
	"bytes"
	"sort"
	"strings"

	"github.com/sivukhin/godjot/tokenizer"
)

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
