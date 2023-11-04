package djot_parser

import (
	"bytes"
	"fmt"
	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/tokenizer"
	"strings"
	"unicode"
)

const RawFormatKey = "RawFormatKey"

type DjotNode int

const (
	DocumentNode = iota
	SectionNode
	ParagraphNode
	HeadingNode
	QuoteNode
	ListItemNode
	ListNode
	CodeNode
	ThematicBreakNode
	DivNode
	PipeTableNode
	ReferenceDefNode
	FootnoteDefNode
	TextNode
	EmphasisNode
	StrongNode
	HighlightedNode
	SubscriptNode
	SuperscriptNode
	InsertNode
	DeleteNode
	SymbolsNode
	VerbatimNode
	LineBreakNode
	LinkNode
	SpanNode
)

func (n DjotNode) String() string {
	switch n {
	case DocumentNode:
		return "DocumentNode"
	case ParagraphNode:
		return "ParagraphNode"
	case HeadingNode:
		return "HeadingNode"
	case QuoteNode:
		return "QuoteNode"
	case ListItemNode:
		return "ListItemNode"
	case ListNode:
		return "ListNode"
	case CodeNode:
		return "CodeNode"
	case ThematicBreakNode:
		return "ThematicBreakNode"
	case DivNode:
		return "DivNode"
	case PipeTableNode:
		return "PipeTableNode"
	case ReferenceDefNode:
		return "ReferenceDefNode"
	case FootnoteDefNode:
		return "FootnoteDefNode"
	case TextNode:
		return "TextNode"
	case EmphasisNode:
		return "EmphasisNode"
	case StrongNode:
		return "StrongNode"
	case HighlightedNode:
		return "HighlightedNode"
	case SubscriptNode:
		return "SubscriptNode"
	case SuperscriptNode:
		return "SuperscriptNode"
	case InsertNode:
		return "InsertNode"
	case DeleteNode:
		return "DeleteNode"
	case SymbolsNode:
		return "SymbolsNode"
	case VerbatimNode:
		return "VerbatimNode"
	case LineBreakNode:
		return "LineBreakNode"
	case LinkNode:
		return "LinkNode"
	default:
		panic(fmt.Errorf("unexpected djot node: %d", n))
	}
}

func convertTokenToNode(token djot_tokenizer.DjotToken) DjotNode {
	switch token {
	case djot_tokenizer.DocumentBlock:
		return DocumentNode
	case djot_tokenizer.HeadingBlock:
		return HeadingNode
	case djot_tokenizer.QuoteBlock:
		return QuoteNode
	case djot_tokenizer.ListItemBlock:
		return ListItemNode
	case djot_tokenizer.CodeBlock:
		return CodeNode
	case djot_tokenizer.DivBlock:
		return DivNode
	case djot_tokenizer.PipeTableBlock:
		return PipeTableNode
	case djot_tokenizer.FootnoteDefBlock:
		return FootnoteDefNode
	case djot_tokenizer.ParagraphBlock:
		return ParagraphNode
	case djot_tokenizer.ThematicBreakToken:
		return ThematicBreakNode
	case djot_tokenizer.EmphasisInline:
		return EmphasisNode
	case djot_tokenizer.StrongInline:
		return StrongNode
	case djot_tokenizer.HighlightedInline:
		return HighlightedNode
	case djot_tokenizer.SubscriptInline:
		return SubscriptNode
	case djot_tokenizer.SuperscriptInline:
		return SuperscriptNode
	case djot_tokenizer.InsertInline:
		return InsertNode
	case djot_tokenizer.DeleteInline:
		return DeleteNode
	}
	return 0
}

func normalizeLinkText(link []byte) []byte { return bytes.ReplaceAll(link, []byte("\n"), nil) }

type DjotContext struct {
	References map[string][]byte
}

func BuildDjotContext(document []byte, list tokenizer.TokenList[djot_tokenizer.DjotToken]) DjotContext {
	context := DjotContext{References: make(map[string][]byte)}
	for i := 0; i < len(list); i++ {
		openToken := list[i]
		if openToken.JumpToPair <= 0 {
			continue
		}
		closeToken := list[i+openToken.JumpToPair]
		switch openToken.Type {
		case djot_tokenizer.ReferenceDefBlock:
			reference := document[openToken.Start:openToken.End]
			reference = bytes.TrimPrefix(reference, []byte(`[`))
			reference = bytes.TrimSuffix(reference, []byte(`]:`))
			link := bytes.Trim(document[openToken.End:closeToken.Start], "\t\r\n ")
			context.References[string(reference)] = link
		}
	}
	return context
}

func isSpaceToken(document []byte, token tokenizer.Token[djot_tokenizer.DjotToken]) bool {
	if token.Type != djot_tokenizer.None && token.Type != djot_tokenizer.SmartSymbolInline {
		return false
	}
	value := document[token.Start:token.End]
	return len(bytes.Trim(value, "\r\n\t ")) == 0
}

type QuoteDirection int

const (
	OpenQuote  QuoteDirection = +1
	CloseQuote                = -1
)

func detectQuoteDirection(document []byte, position int) QuoteDirection {
	if document[position] == '{' {
		return OpenQuote
	}
	if position+1 < len(document) && document[position+1] == '}' {
		return CloseQuote
	}
	if position == 0 {
		return OpenQuote
	}
	if position == len(document)-1 {
		return CloseQuote
	}
	if unicode.IsSpace(rune(document[position-1])) {
		return OpenQuote
	}
	if unicode.IsSpace(rune(document[position+1])) {
		return CloseQuote
	}
	if unicode.IsPunct(rune(document[position-1])) {
		return OpenQuote
	}
	if unicode.IsPunct(rune(document[position+1])) {
		return CloseQuote
	}
	return OpenQuote
}

func trimPadding(document []byte, list tokenizer.TokenList[djot_tokenizer.DjotToken]) tokenizer.TokenList[djot_tokenizer.DjotToken] {
	start, end := 0, len(list)
	for start < end && isSpaceToken(document, list[start]) {
		start++
	}
	for start < end && isSpaceToken(document, list[end-1]) {
		end--
	}
	if start < end {
		return list[start:end]
	}
	return nil
}

func buildDjotAst(
	document []byte,
	context DjotContext,
	list tokenizer.TokenList[djot_tokenizer.DjotToken],
	textNode bool,
) []Tree[DjotNode] {
	if len(list) == 0 {
		return nil
	}
	nodes := make([]Tree[DjotNode], 0)
	i := 0
	for i < len(list) {
		openToken := list[i]
		textBytes := document[openToken.Start:openToken.End]
		closeToken := list[i+openToken.JumpToPair]
		attributes := make(map[string]string)
		nextI := i + openToken.JumpToPair + 1
		for nextI < len(list) && list[nextI].Type == djot_tokenizer.Attribute {
			for key, value := range list[nextI].Attributes {
				attributes[key] = value
			}
			nextI++
		}
		switch openToken.Type {
		case
			djot_tokenizer.DocumentBlock,
			djot_tokenizer.QuoteBlock,
			djot_tokenizer.CodeBlock,
			djot_tokenizer.DivBlock,
			djot_tokenizer.ThematicBreakToken,
			djot_tokenizer.ParagraphBlock,
			djot_tokenizer.EmphasisInline,
			djot_tokenizer.StrongInline,
			djot_tokenizer.HighlightedInline,
			djot_tokenizer.SubscriptInline,
			djot_tokenizer.SuperscriptInline,
			djot_tokenizer.InsertInline,
			djot_tokenizer.DeleteInline:
			nodes = append(nodes, Tree[DjotNode]{
				Type: convertTokenToNode(openToken.Type),
				Children: buildDjotAst(
					document,
					context,
					trimPadding(document, list[i+1:i+openToken.JumpToPair]),
					textNode ||
						openToken.Type == djot_tokenizer.ParagraphBlock ||
						openToken.Type == djot_tokenizer.HeadingBlock,
				),
				Token:      textBytes,
				Attributes: openToken.Attributes,
			})
		case djot_tokenizer.SymbolsInline:
			nodes = append(nodes, Tree[DjotNode]{
				Type: SymbolsNode,
				Text: document[openToken.End:closeToken.Start],
			})
		case djot_tokenizer.AutolinkInline:
			link := normalizeLinkText(document[openToken.End:closeToken.Start])
			nodes = append(nodes, Tree[DjotNode]{
				Type:     LinkNode,
				Text:     link,
				Children: []Tree[DjotNode]{{Type: TextNode, Text: link}},
			})
		case djot_tokenizer.VerbatimInline:
			text := document[openToken.End:list[i+openToken.JumpToPair].Start]
			if trimmed := bytes.Trim(text, " "); bytes.HasPrefix(trimmed, []byte("`")) && bytes.HasSuffix(trimmed, []byte("`")) {
				text = text[1 : len(text)-1]
			}
			attributes := openToken.Attributes
			if nextI < len(list) && list[nextI].Type == djot_tokenizer.RawFormatInline {
				if attributes == nil {
					attributes = make(map[string]string)
				}
				rawFormatOpen := list[nextI]
				rawFormatClose := list[nextI+rawFormatOpen.JumpToPair]
				attributes[RawFormatKey] = string(document[rawFormatOpen.End:rawFormatClose.Start])
				nextI += rawFormatOpen.JumpToPair + 1
			}
			nodes = append(nodes, Tree[DjotNode]{
				Type:       VerbatimNode,
				Token:      textBytes,
				Text:       text,
				Attributes: attributes,
			})
		case djot_tokenizer.SpanInline, djot_tokenizer.ImageSpanInline:
			if nextI < len(list) {
				nextToken := list[nextI]
				if nextToken.Type == djot_tokenizer.LinkUrlInline {
					nodes = append(nodes, Tree[DjotNode]{
						Type:     LinkNode,
						Text:     normalizeLinkText(document[nextToken.End:list[nextI+nextToken.JumpToPair].Start]),
						Children: buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode),
					})
					nextI += nextToken.JumpToPair + 1
				} else if nextToken.Type == djot_tokenizer.LinkReferenceInline {
					reference := string(normalizeLinkText(document[nextToken.End:list[nextI+nextToken.JumpToPair].Start]))
					//if len(reference) == 0 {
					//	reference =
					//}
					nodes = append(nodes, Tree[DjotNode]{
						Type:     LinkNode,
						Text:     normalizeLinkText(context.References[reference]),
						Children: buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode),
					})
					nextI += nextToken.JumpToPair + 1
				} else if len(attributes) > 0 {
					nodes = append(nodes, Tree[DjotNode]{
						Type:       SpanNode,
						Children:   buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode),
						Attributes: attributes,
					})
				} else {
					nodes = append(nodes, Tree[DjotNode]{
						Type: TextNode,
						Text: textBytes,
					})
					nodes = append(nodes, buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode)...)
					nodes = append(nodes, Tree[DjotNode]{
						Type: TextNode,
						Text: document[closeToken.Start:closeToken.End],
					})
				}
			} else if nextI >= len(list) {
				nodes = append(nodes, Tree[DjotNode]{
					Type: TextNode,
					Text: textBytes,
				})
				nodes = append(nodes, buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode)...)
				nodes = append(nodes, Tree[DjotNode]{
					Type: TextNode,
					Text: document[closeToken.Start:closeToken.End],
				})
			}
		case djot_tokenizer.EscapedSymbolInline:
			if textNode {
				text := textBytes
				if text[len(text)-1] == '\n' {
					nodes = append(nodes, Tree[DjotNode]{Type: LineBreakNode}, Tree[DjotNode]{Type: TextNode, Text: []byte("\n")})
				} else {
					nodes = append(nodes, Tree[DjotNode]{Type: TextNode, Text: text[1:]})
				}
			}
		case djot_tokenizer.SmartSymbolInline:
			textString := strings.Trim(string(textBytes), "{}")
			if textNode {
				quoteDirection := detectQuoteDirection(document, openToken.Start)
				if openToken.Type == djot_tokenizer.SmartSymbolInline {
					if textString == "\"" && quoteDirection == OpenQuote {
						textBytes = []byte(`&ldquo;`)
					} else if textString == "\"" && quoteDirection == CloseQuote {
						textBytes = []byte(`&rdquo;`)
					} else if textString == "'" && quoteDirection == OpenQuote {
						textBytes = []byte(`&lsquo;`)
					} else if textString == "'" && quoteDirection == CloseQuote {
						textBytes = []byte(`&rsquo;`)
					} else if textString == "..." {
						textBytes = []byte(`&hellip;`)
					} else if strings.Count(textString, "-") == len(textString) {
						if len(textString)%3 == 0 {
							textBytes = bytes.Repeat([]byte(`&mdash;`), len(textString)/3)
						} else if len(textString)%2 == 0 {
							textBytes = bytes.Repeat([]byte(`&ndash;`), len(textString)/2)
						} else {
							textBytes = append(bytes.Repeat([]byte(`&ndash;`), (len(textString)-3)/2), []byte(`&mdash;`)...)
						}
					}
				}
				nodes = append(nodes, Tree[DjotNode]{Type: TextNode, Text: textBytes})
			}
		case djot_tokenizer.None:
			if textNode {
				if len(attributes) > 0 {
					split := bytes.LastIndexByte(textBytes, ' ')
					nodes = append(nodes, Tree[DjotNode]{Type: TextNode, Text: textBytes[:split+1]})
					nodes = append(nodes, Tree[DjotNode]{
						Type:       SpanNode,
						Attributes: attributes,
						Children:   []Tree[DjotNode]{{Type: TextNode, Text: textBytes[split+1:]}},
					})
				} else {
					nodes = append(nodes, Tree[DjotNode]{Type: TextNode, Text: textBytes})
				}
			}
		}
		i = nextI
	}
	return nodes
}
