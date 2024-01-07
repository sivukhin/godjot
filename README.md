## godjot

[Djot](https://djot.net/) parser implemented in Go language

**godjot** provides API to parse AST from djot string 
``` go
var djot []byte
ast := djot_parser.BuildDjotAst(djot)
```

AST is loosely typed and described with following simple struct:
```go
type TreeNode[T ~int] struct {
    Type       T                     // one of DjotNode options
    Attributes *tokenizer.Attributes // string attributes of node
    Children   []TreeNode[T]         // list of child
    Text       []byte                // content of text nodes: TextNode / SymbolsNode / VerbatimNode
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
).ConvertDjotToHtml(ast...)
```