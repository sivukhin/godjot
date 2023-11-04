package htmlwriter

import "strings"

type HtmlWriter struct {
	Builder strings.Builder
}

type HtmlAttribute struct {
	Key, Value string
}

func (w *HtmlWriter) String() string { return w.Builder.String() }

func (w *HtmlWriter) Tag(tag string, attributes ...HtmlAttribute) {
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

func (w *HtmlWriter) InTag(tag string, attributes ...HtmlAttribute) func(func()) {
	return func(content func()) {
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
		content()
		w.Builder.WriteString("</")
		w.Builder.WriteString(tag)
		w.Builder.WriteString(">")
	}
}

func (w *HtmlWriter) WriteBytes(text []byte)  { w.Builder.Write(text) }
func (w *HtmlWriter) WriteString(text string) { w.Builder.WriteString(text) }
