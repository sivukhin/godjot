package parser

import (
	"github.com/sivukhin/godjot/tokenizer"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func printDjotAst(builder *strings.Builder, trees ...Tree[DjotNode]) {
	for _, tree := range trees {
		switch tree.Type {
		case InsertNode:
			builder.WriteString("<ins>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</ins>")
		case DeleteNode:
			builder.WriteString("<del>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</del>")
		case SuperscriptNode:
			builder.WriteString("<sup>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</sup>")
		case SubscriptNode:
			builder.WriteString("<sub>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</sub>")
		case HighlightedNode:
			builder.WriteString("<mark>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</mark>")
		case EmphasisNode:
			builder.WriteString("<em>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</em>")
		case StrongNode:
			builder.WriteString("<strong>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</strong>")
		case DocumentNode:
			builder.WriteString("<p>")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</p>\n")
		case LinkNode:
			builder.WriteString("<a href=\"")
			builder.WriteString(tree.Attributes["Url"])
			builder.WriteString("\">")
			printDjotAst(builder, tree.Children...)
			builder.WriteString("</a>")
		case AutolinkNode:
			builder.WriteString("<a href=\"")
			link := tree.Text
			if strings.Contains(link, "@") {
				link = "mailto:" + link
			}
			builder.WriteString(link)
			builder.WriteString("\">")
			builder.WriteString(tree.Text)
			builder.WriteString("</a>")
		case VerbatimNode:
			builder.WriteString("<code>")
			builder.WriteString(tree.Text)
			builder.WriteString("</code>")
		case VerbatimInlineMathNode:
			builder.WriteString("<span class=\"math inline\">\\(")
			builder.WriteString(tree.Text)
			builder.WriteString("\\)</span>")
		case VerbatimDisplayMathNode:
			builder.WriteString("<span class=\"math display\">\\[")
			builder.WriteString(tree.Text)
			builder.WriteString("\\]</span>")
		case LineBreakNode:
			builder.WriteString("<br>\n")
		case TextNode:
			builder.WriteString(tree.Text)
		}
	}
}

func printDjot(text string) string {
	var builder strings.Builder
	printDjotAst(&builder, DjotAst([]byte(text)))
	return builder.String()
}

func TestDjotInlineText(t *testing.T) {
	for _, tt := range []struct{ djot, html string }{
		{
			djot: "[My link text](http://example.com)",
			html: "<p><a href=\"http://example.com\">My link text</a></p>\n",
		},
		{
			djot: "[My link text](http://example.com?product_number=234234234234\n234234234234)",
			html: "<p><a href=\"http://example.com?product_number=234234234234234234234234\">My link text</a></p>\n",
		},
		{
			djot: "<https://pandoc.org/lua-filters>\n<me@example.com>",
			html: "<p><a href=\"https://pandoc.org/lua-filters\">https://pandoc.org/lua-filters</a>\n<a href=\"mailto:me@example.com\">me@example.com</a></p>\n",
		},
		{
			djot: "``Verbatim with a backtick` character``\n`Verbatim with three backticks ``` character`",
			html: "<p><code>Verbatim with a backtick` character</code>\n<code>Verbatim with three backticks ``` character</code></p>\n",
		},
		{
			djot: "`` `foo` ``",
			html: "<p><code>`foo`</code></p>\n",
		},
		{
			djot: "_emphasized text_\n*strong emphasis*",
			html: "<p><em>emphasized text</em>\n<strong>strong emphasis</strong></p>\n",
		},
		{
			djot: "_ Not emphasized (spaces). _\n___ (not an emphasized `_` character)",
			html: "<p>_ Not emphasized (spaces). _\n___ (not an emphasized <code>_</code> character)</p>\n",
		},
		{
			djot: "__emphasis inside_ emphasis_",
			html: "<p><em><em>emphasis inside</em> emphasis</em></p>\n",
		},
		{
			djot: "{_ this is emphasized, despite the spaces! _}",
			html: "<p><em> this is emphasized, despite the spaces! </em></p>\n",
		},
		{
			djot: "This is {=highlighted text=}.",
			html: "<p>This is <mark>highlighted text</mark>.</p>\n",
		},
		{
			djot: "H~2~O and djot^TM^",
			html: "<p>H<sub>2</sub>O and djot<sup>TM</sup></p>\n",
		},
		{
			djot: "H{~one two buckle my shoe~}O",
			html: "<p>H<sub>one two buckle my shoe</sub>O</p>\n",
		},
		{
			djot: "My boss is {-mean-}{+nice+}.",
			html: "<p>My boss is <del>mean</del><ins>nice</ins>.</p>\n",
		},
		{
			djot: "Einstein derived $`e=mc^2`.\nPythagoras proved\n$$` x^n + y^n = z^n `",
			html: "<p>Einstein derived <span class=\"math inline\">\\(e=mc^2\\)</span>.\nPythagoras proved\n<span class=\"math display\">\\[ x^n + y^n = z^n \\]</span></p>\n",
		},
		{
			djot: "This is a soft\nbreak and this is a hard\\\nbreak.",
			html: "<p>This is a soft\nbreak and this is a hard<br>\nbreak.</p>\n",
		},
	} {
		t.Run(tt.html, func(t *testing.T) {
			require.Equalf(
				t, tt.html, printDjot(tt.djot),
				"invalid html (%v != %v), djot tokens: %v",
				tt.html, printDjot(tt.djot),
				tokenizer.DjotTokens([]byte(tt.djot)),
			)
		})
	}
}
