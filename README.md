## godjot

[![Go Report Card](https://goreportcard.com/badge/git.sr.ht/~ser/godjot)](https://goreportcard.com/report/git.sr.ht/~ser/godjot)
[![Build status](https://builds.sr.ht/~ser/godjot/.build.yml.svg)](https://builds.sr.ht/~ser/godjot/.build.yml?)

[Djot](https://github.com/jgm/djot) markup language parser implemented in Go language

## This is a fork

The [original project](https://github.com/sivukhin/godjot) is on github. I
submit PRs upstream; you may want to check that project first. I use this
library in a few different projects, and depend on some of the changes.

### Installation

You can install **godjot** as a standalone binary:
```shell
$> go install git.sr.ht/~ser/godjot/v2@latest
$> echo '*Hello*, _world_' | godjot
<p><strong>Hello</strong>, <em>world</em></p>
```

### Usage

**godjot** provides API to parse AST from djot string 
``` go
var djot []byte
ast := djot_parser.BuildDjotAst(djot)
```

AST is loosely typed and described with following simple struct:
```go
type TreeNode[T ~int] struct {
    Type       T                     // one of DjotNode options
    Attributes tokenizer.Attributes  // string attributes of node
    Children   []TreeNode[T]         // list of child
    Text       []byte                // not nil only for TextNode
}
```

You can transform AST to HTML with predefined set of rules:
```go
content := djot_html.New().ConvertDjot(&djot_html.HtmlWriter{}, ast...).String()
```

Or, you can override some default conversion rules:
```go
content := djot_html.New(
    djot_html.DefaultConversionRegistry,
    map[djot_parser.DjotNode]djot_parser.Conversion[*djot_html.HtmlWriter]{
        djot_parser.ImageNode: func(state djot_parser.ConversionState[*djot_html.HtmlWriter], next func(c djot_parser.Children)) {
            state.Writer.
                OpenTag("figure").
                OpenTag("img", state.Node.Attributes.Entries()...).
                OpenTag("figcaption").
                WriteString(state.Node.Attributes.Get(djot_parser.ImgAltKey)).
                CloseTag("figcaption").
                CloseTag("figure")
        },
    }
).ConvertDjot(&djot_html.HtmlWriter{}, ast...).String()
```

This implementation passes all examples provided in the [spec](https://htmlpreview.github.io/?https://github.com/jgm/djot/blob/master/doc/syntax.html) but can diverge from original javascript implementation in some cases.

### Reproducing reported bugs

Include the Djot input when reporting a rendering bug. Maintainers can check
the expected HTML explicitly or compare godjot with the pinned Djot reference
renderer. The [issue reproducer guide](scripts/issue-repro/README.md) shows
how to run either check locally and in nightly CI.
