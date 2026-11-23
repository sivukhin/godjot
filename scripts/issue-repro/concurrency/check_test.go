//go:build issuerepro && race

package concurrency_test

import (
	"maps"
	"testing"

	"github.com/sivukhin/godjot/v2/djot_html"
	"github.com/sivukhin/godjot/v2/djot_parser"
	"github.com/sivukhin/godjot/v2/djot_tokenizer"
)

func TestParserState(t *testing.T) {
	before := maps.Clone(djot_tokenizer.StartSymbols)
	djot_parser.BuildDjotAst([]byte("An _ordinary_ paragraph.\n"))
	if !maps.Equal(before, djot_tokenizer.StartSymbols) {
		t.Error("parsing a document changed the shared set of inline start symbols")
	}
}

func TestParallelRender(t *testing.T) {
	renderer := djot_html.New()
	for _, example := range []struct{ name, input, html string }{
		{"emphasis", "An _ordinary_ paragraph.\n", "<p>An <em>ordinary</em> paragraph.</p>\n"},
		{"strong", "A *second* document.\n", "<p>A <strong>second</strong> document.</p>\n"},
		{"code", "A `third` example.\n", "<p>A <code>third</code> example.</p>\n"},
	} {
		// Parallel subtests resume after this loop. Each has its own parsed
		// document and writer, while all of them read the same renderer.
		doc := djot_parser.BuildDjotAst([]byte(example.input))
		t.Run(example.name, func(t *testing.T) {
			t.Parallel()
			got := renderer.ConvertDjot(&djot_html.HtmlWriter{}, doc...).String()
			if got != example.html {
				t.Errorf("rendered %q, want %q", got, example.html)
			}
		})
	}
}
