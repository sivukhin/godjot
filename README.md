## godjot

[![Go Report Card][go-report-image]][go-report-url]

[go-report-image]: https://goreportcard.com/badge/github.com/sivukhin/godjot
[go-report-url]: https://goreportcard.com/report/github.com/sivukhin/godjot

[Djot](https://github.com/jgm/djot) markup language parser implemented in Go language

### Installation

You can install **godjot** as a standalone binary:
```shell
$> go install github.com/sivukhin/godjot@latest
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
content := djot_parser.NewConversionContext(
    "html", 
    djot_parser.DefaultConversionRegistry,
    map[djot_parser.DjotNode]djot_parser.Conversion{
        /*
            You can overwrite default conversion rules with custom map
            djot_parser.ImageNode: func(state djot_parser.ConversionState, next func(c djot_parser.Children)) {
                state.Writer.
                    OpenTag("figure").
                    OpenTag("img", state.Node.Attributes.Entries()...).
                    OpenTag("figcaption").
                    WriteString(state.Node.Attributes.Get(djot_parser.ImgAltKey)).
                    CloseTag("figcaption").
                    CloseTag("figure")
            }
        */
    }
).ConvertDjotToHtml(&html_writer.HtmlWriter{}, ast...)
```

This implementation passes all examples provided in the [spec](https://htmlpreview.github.io/?https://github.com/jgm/djot/blob/master/doc/syntax.html) but can diverge from original javascript implementation in some cases.
