package djot_parser

import (
	"bytes"
	"fmt"
	"github.com/sivukhin/godjot/djot_tokenizer"
	"github.com/sivukhin/godjot/tokenizer"
	"strings"
	"unicode"
)

const (
	RawFormatKey    = "$RawFormatKey"
	HeadingLevelKey = "$HeadingLevelKey"

	IdKey       = "id"
	RoleKey     = "role"
	LinkHrefKey = "href"
	ImgAltKey   = "alt"
	ImgSrcKey   = "src"
)

type DjotNode int

const (
	DocumentNode DjotNode = iota
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
	ImageNode
	SpanNode
)

func (n DjotNode) String() string {
	switch n {
	case DocumentNode:
		return "DocumentNode"
	case SectionNode:
		return "SectionNode"
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
	FootnoteId map[string]int
}

func BuildDjotContext(document []byte, list tokenizer.TokenList[djot_tokenizer.DjotToken]) DjotContext {
	context := DjotContext{
		References: make(map[string][]byte),
		FootnoteId: make(map[string]int),
	}

	footnoteId := 1

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
		case djot_tokenizer.FootnoteDefBlock:
			reference := document[openToken.Start:openToken.End]
			reference = bytes.TrimPrefix(reference, []byte(`[^`))
			reference = bytes.TrimSuffix(reference, []byte(`]:`))
			context.FootnoteId[string(reference)] = footnoteId
			footnoteId++
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

func selectText(document []byte, list tokenizer.TokenList[djot_tokenizer.DjotToken]) []byte {
	text := make([]byte, 0)
	for _, token := range list {
		if token.Type == djot_tokenizer.None || token.Type == djot_tokenizer.SmartSymbolInline {
			text = append(text, document[token.Start:token.End]...)
		}
	}
	return text
}

func createSectionId(s string) string {
	id := strings.Builder{}
	hasDash := false
	for _, c := range s {
		if unicode.IsSpace(c) || c == '\n' || c == '\t' {
			hasDash = true
		} else if unicode.IsPunct(c) {
			continue
		} else {
			if hasDash {
				id.WriteString("-")
			}
			hasDash = false
			id.WriteRune(c)
		}
	}
	return id.String()
}

func buildDjotAst(
	document []byte,
	context DjotContext,
	list tokenizer.TokenList[djot_tokenizer.DjotToken],
	textNode bool,
) []TreeNode[DjotNode] {
	if len(list) == 0 {
		return nil
	}
	sectionPop := make(map[int]int)
	{
		sectionStart, sectionLevel := make([]int, 0), make([]int, 0)
		i := 0
		for i < len(list) {
			openToken := list[i]
			switch openToken.Type {
			case djot_tokenizer.HeadingBlock:
				level := len(bytes.TrimSuffix(document[openToken.Start:openToken.End], []byte(" ")))
				pop := 0
				for len(sectionLevel) > 0 && sectionLevel[len(sectionLevel)-1] >= level {
					sectionStart = sectionStart[:len(sectionStart)-1]
					sectionLevel = sectionLevel[:len(sectionLevel)-1]
					pop++
				}
				sectionPop[i] = pop
			}
			i += openToken.JumpToPair + 1
		}
		for len(sectionLevel) > 0 {
			sectionPop[sectionStart[len(sectionStart)-1]] = len(list)
			sectionStart = sectionStart[:len(sectionStart)-1]
			sectionLevel = sectionLevel[:len(sectionLevel)-1]
		}
	}

	nodes := make([]TreeNode[DjotNode], 0)
	sections := []*[]TreeNode[DjotNode]{&nodes}
	nodesRef := &nodes
	{
		i := 0
		for i < len(list) {
			openToken := list[i]
			textBytes := document[openToken.Start:openToken.End]
			closeToken := list[i+openToken.JumpToPair]
			nextI := i + openToken.JumpToPair + 1
			attributes := (&tokenizer.Attributes{}).MergeWith(openToken.Attributes)
			for nextI < len(list) && list[nextI].Type == djot_tokenizer.Attribute {
				attributes.MergeWith(list[nextI].Attributes)
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
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type: convertTokenToNode(openToken.Type),
					Children: buildDjotAst(
						document,
						context,
						trimPadding(document, list[i+1:i+openToken.JumpToPair]),
						textNode ||
							openToken.Type == djot_tokenizer.ParagraphBlock ||
							openToken.Type == djot_tokenizer.HeadingBlock ||
							openToken.Type == djot_tokenizer.CodeBlock,
					),
					Attributes: attributes,
				})
			case djot_tokenizer.HeadingBlock:
				if openToken.Type == djot_tokenizer.HeadingBlock {
					attributes.Set(HeadingLevelKey, string(bytes.TrimSuffix(document[openToken.Start:openToken.End], []byte(" "))))
				}
				sections = sections[:len(sections)-sectionPop[i]]
				nodesRef = sections[len(sections)-1]
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type:       SectionNode,
					Attributes: (&tokenizer.Attributes{}).Set("id", createSectionId(string(selectText(document, list[i+1:i+openToken.JumpToPair])))),
				})
				nodesRef = &(*nodesRef)[len(*nodesRef)-1].Children
				sections = append(sections, nodesRef)
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type: convertTokenToNode(openToken.Type),
					Children: buildDjotAst(
						document,
						context,
						trimPadding(document, list[i+1:i+openToken.JumpToPair]),
						textNode ||
							openToken.Type == djot_tokenizer.ParagraphBlock ||
							openToken.Type == djot_tokenizer.HeadingBlock ||
							openToken.Type == djot_tokenizer.CodeBlock,
					),
					Attributes: attributes,
				})
			case djot_tokenizer.SymbolsInline:
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type:       SymbolsNode,
					Text:       document[openToken.End:closeToken.Start],
					Attributes: attributes,
				})
			case djot_tokenizer.AutolinkInline:
				link := normalizeLinkText(document[openToken.End:closeToken.Start])
				href := string(link)
				if strings.Contains(href, "@") {
					href = "mailto:" + href
				}
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type:       LinkNode,
					Children:   []TreeNode[DjotNode]{{Type: TextNode, Text: link}},
					Attributes: attributes.Set(LinkHrefKey, href),
				})
			case djot_tokenizer.VerbatimInline:
				text := document[openToken.End:list[i+openToken.JumpToPair].Start]
				if trimmed := bytes.Trim(text, " "); bytes.HasPrefix(trimmed, []byte("`")) && bytes.HasSuffix(trimmed, []byte("`")) {
					text = text[1 : len(text)-1]
				}
				if nextI < len(list) && list[nextI].Type == djot_tokenizer.RawFormatInline {
					rawFormatOpen := list[nextI]
					rawFormatClose := list[nextI+rawFormatOpen.JumpToPair]
					attributes.Set(RawFormatKey, string(document[rawFormatOpen.End:rawFormatClose.Start]))
					nextI += rawFormatOpen.JumpToPair + 1
				}
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type:       VerbatimNode,
					Text:       text,
					Attributes: attributes,
				})
			case djot_tokenizer.FootnoteReferenceInline:
				footnoteId := context.FootnoteId[string(document[openToken.End:closeToken.Start])]
				*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
					Type:     LinkNode,
					Text:     []byte(fmt.Sprintf("#fn%v", footnoteId)),
					Children: []TreeNode[DjotNode]{{Type: SuperscriptNode, Children: []TreeNode[DjotNode]{{Type: TextNode, Text: []byte(fmt.Sprintf("%v", footnoteId))}}}},
					Attributes: attributes.
						Set(IdKey, fmt.Sprintf("fnref%v", footnoteId)).
						Set(LinkHrefKey, fmt.Sprintf("#fn%v", footnoteId)).
						Set(RoleKey, "doc-noteref"),
				})
			case djot_tokenizer.ImageSpanInline:
				var nextToken tokenizer.Token[djot_tokenizer.DjotToken]
				if nextI < len(list) {
					nextToken = list[nextI]
				}
				if nextToken.Type == djot_tokenizer.LinkUrlInline {
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type: ImageNode,
						Attributes: attributes.
							Set(ImgAltKey, string(selectText(document, list[i+1:i+openToken.JumpToPair]))).
							Set(ImgSrcKey, string(normalizeLinkText(document[nextToken.End:list[nextI+nextToken.JumpToPair].Start]))),
					})
					nextI += nextToken.JumpToPair + 1
				} else if nextToken.Type == djot_tokenizer.LinkReferenceInline {
					reference := normalizeLinkText(document[nextToken.End:list[nextI+nextToken.JumpToPair].Start])
					if len(reference) == 0 {
						reference = selectText(document, list[i+1:i+openToken.JumpToPair])
					}
					attributes.Set(ImgAltKey, string(selectText(document, list[i+1:i+openToken.JumpToPair])))
					if href := string(normalizeLinkText(context.References[string(reference)])); href != "" {
						attributes.Set(ImgSrcKey, href)
					}
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type:       ImageNode,
						Attributes: attributes,
					})
					nextI += nextToken.JumpToPair + 1
				} else {
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type: TextNode,
						Text: textBytes,
					})
					*nodesRef = append(*nodesRef, buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode)...)
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type: TextNode,
						Text: document[closeToken.Start:closeToken.End],
					})
				}
			case djot_tokenizer.SpanInline:
				var nextToken tokenizer.Token[djot_tokenizer.DjotToken]
				if nextI < len(list) {
					nextToken = list[nextI]
				}
				if nextToken.Type == djot_tokenizer.LinkUrlInline {
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type:       LinkNode,
						Children:   buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode),
						Attributes: attributes.Set(LinkHrefKey, string(normalizeLinkText(document[nextToken.End:list[nextI+nextToken.JumpToPair].Start]))),
					})
					nextI += nextToken.JumpToPair + 1
				} else if nextToken.Type == djot_tokenizer.LinkReferenceInline {
					reference := normalizeLinkText(document[nextToken.End:list[nextI+nextToken.JumpToPair].Start])
					if len(reference) == 0 {
						reference = selectText(document, list[i+1:i+openToken.JumpToPair])
					}
					if href := string(normalizeLinkText(context.References[string(reference)])); href != "" {
						attributes.Set(LinkHrefKey, href)
					}
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type:       LinkNode,
						Attributes: attributes,
						Children:   buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode),
					})
					nextI += nextToken.JumpToPair + 1
				} else if attributes.Size() > 0 {
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type:       SpanNode,
						Children:   buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode),
						Attributes: attributes,
					})
				} else {
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type: TextNode,
						Text: textBytes,
					})
					*nodesRef = append(*nodesRef, buildDjotAst(document, context, list[i+1:i+openToken.JumpToPair], textNode)...)
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
						Type: TextNode,
						Text: document[closeToken.Start:closeToken.End],
					})
				}
			case djot_tokenizer.EscapedSymbolInline:
				if textNode {
					text := textBytes
					if text[len(text)-1] == '\n' {
						*nodesRef = append(*nodesRef, TreeNode[DjotNode]{Type: LineBreakNode}, TreeNode[DjotNode]{Type: TextNode, Text: []byte("\n")})
					} else {
						*nodesRef = append(*nodesRef, TreeNode[DjotNode]{Type: TextNode, Text: text[1:]})
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
					*nodesRef = append(*nodesRef, TreeNode[DjotNode]{Type: TextNode, Text: textBytes})
				}
			case djot_tokenizer.None:
				if textNode {
					if attributes.Size() > 0 {
						split := bytes.LastIndexByte(textBytes, ' ')
						*nodesRef = append(*nodesRef, TreeNode[DjotNode]{Type: TextNode, Text: textBytes[:split+1]})
						*nodesRef = append(*nodesRef, TreeNode[DjotNode]{
							Type:       SpanNode,
							Attributes: attributes,
							Children:   []TreeNode[DjotNode]{{Type: TextNode, Text: textBytes[split+1:]}},
						})
					} else {
						*nodesRef = append(*nodesRef, TreeNode[DjotNode]{Type: TextNode, Text: textBytes})
					}
				}
			}
			i = nextI
		}
	}
	return nodes
}
