package parser

import (
	"bytes"
	"fmt"
	"github.com/sivukhin/godjot/tokenizer"
)

type DjotNode int

const (
	TextNode = iota
	EmphasisNode
	StrongNode
	HighlightedNode
	SubscriptNode
	SuperscriptNode
	InsertNode
	DeleteNode
	AutolinkNode
	SymbolsNode
	VerbatimNode
	VerbatimInlineMathNode
	VerbatimDisplayMathNode
	LineBreakNode
	DocumentNode
	LinkNode
)

type MatchContext struct {
	*TokenReader[tokenizer.DjotToken]
	Tree *[]Tree[DjotNode]
}

func MatchDjotInlineNode(
	context MatchContext,
	current TokenPosition,
	next *TokenPosition,
	token tokenizer.DjotToken,
	node DjotNode,
) bool {
	if _, match := context.MatchToken(current, next, token); !match {
		return false
	}
	if !MatchDjotAst(context, *next, next) {
		panic(fmt.Errorf("invalid djot sequence at position %v", *next))
	}
	if _, match := context.MatchToken(*next, next, token^tokenizer.Open); !match {
		panic(fmt.Errorf("unbalanced djot sequence at position %v", *next))
	}
	*context.Tree = []Tree[DjotNode]{{Type: node, Children: *context.Tree}}
	return true
}

func MatchDjotTextNode(
	context MatchContext,
	current TokenPosition,
	next *TokenPosition,
	token tokenizer.DjotToken,
	node DjotNode,
) bool {
	if _, match := context.MatchToken(current, next, token); !match {
		return false
	}
	firstTextToken, lastTextToken := *next, *next
	for {
		if peek, ok := context.Peek(*next, 0); !ok || peek.Type == token^tokenizer.Open {
			break
		}
		lastTextToken = *next
		*next++
	}
	if _, match := context.MatchToken(*next, next, token^tokenizer.Open); !match {
		panic(fmt.Errorf("invalid djot sequence at position %v (token=%v)", current+1+*next, context.Tokens[*next]))
	}
	text := string(context.Text[context.Tokens[firstTextToken].Start:context.Tokens[lastTextToken].End])
	*context.Tree = []Tree[DjotNode]{{Type: node, Text: text}}
	return true
}

func MatchDjotVerbatim(context MatchContext, current TokenPosition, next *TokenPosition) bool {
	verbatimOpenToken, match := context.MatchToken(current, next, tokenizer.Verbatim)
	if !match {
		return false
	}
	firstTextToken, lastTextToken := *next, *next
	for {
		if peek, ok := context.Peek(*next, 0); !ok || peek.Type == tokenizer.Verbatim^tokenizer.Open {
			break
		}
		lastTextToken = *next
		*next++
	}
	if _, match := context.MatchToken(*next, next, tokenizer.Verbatim^tokenizer.Open); !match {
		panic(fmt.Errorf("invalid djot sequence at position %v (token=%v)", current+1+*next, context.Tokens[*next]))
	}
	text := context.Text[context.Tokens[firstTextToken].Start:context.Tokens[lastTextToken].End]
	trimmed := bytes.TrimRight(bytes.TrimLeft(text, " "), " ")
	if bytes.HasPrefix(trimmed, []byte("`")) && bytes.HasSuffix(trimmed, []byte("`")) {
		text = text[1 : len(text)-1]
	}
	verbatimOpen := context.Text[verbatimOpenToken.Start:verbatimOpenToken.End]
	var node DjotNode = VerbatimNode
	if bytes.HasPrefix(verbatimOpen, []byte("$$")) {
		node = VerbatimDisplayMathNode
	} else if bytes.HasPrefix(verbatimOpen, []byte("$")) {
		node = VerbatimInlineMathNode
	}
	*context.Tree = []Tree[DjotNode]{{Type: node, Text: string(text)}}
	return true
}

func MatchDjotLink(context MatchContext, current TokenPosition, next *TokenPosition) bool {
	spanOpen, ok := context.MatchToken(current, next, tokenizer.Span)
	if !ok {
		return false
	}
	linkTypeOpen, ok := context.Peek(current, spanOpen.JumpToPair+1)
	if !ok {
		return false
	}
	if linkTypeOpen.Type == tokenizer.LinkUrl {
		linkTypeClose, ok := context.Peek(current, spanOpen.JumpToPair+1+linkTypeOpen.JumpToPair)
		if !ok {
			return false
		}
		urlContext := MatchContext{
			TokenReader: context.TokenReader.Span(current+1, current+spanOpen.JumpToPair),
			Tree:        &[]Tree[DjotNode]{},
		}
		if !MatchDjotAst(urlContext, 0, next) {
			panic(fmt.Errorf("invalid djot sequence at position %v (token=%v)", current+1+*next, urlContext.Tokens[*next]))
		}
		*context.Tree = []Tree[DjotNode]{{Type: LinkNode, Children: *urlContext.Tree, Attributes: map[string]string{
			"Url": string(context.Text[linkTypeOpen.End:linkTypeClose.Start]),
		}}}
		*next = current + spanOpen.JumpToPair + 1 + linkTypeOpen.JumpToPair + 1
		return true
	}
	return false
}

func MatchDjotEscapedSymbol(context MatchContext, current TokenPosition, next *TokenPosition) bool {
	token, ok := context.MatchToken(current, next, tokenizer.Escaped)
	if !ok {
		return false
	}
	if context.Text[token.End-1] == '\n' {
		*context.Tree = []Tree[DjotNode]{{Type: LineBreakNode}}
	} else {
		*context.Tree = []Tree[DjotNode]{{Type: TextNode, Text: string(context.Text[token.Start+1 : token.End])}}
	}
	return true
}

func MatchDjotText(context MatchContext, current TokenPosition, next *TokenPosition) bool {
	token, ok := context.MatchToken(current, next, tokenizer.None)
	if !ok {
		return false
	}
	*context.Tree = []Tree[DjotNode]{{Type: TextNode, Text: string(context.Text[token.Start:token.End])}}
	return true
}

func MatchDjotAst(context MatchContext, current TokenPosition, next *TokenPosition) bool {
	level := make([]Tree[DjotNode], 0)
	//goland:noinspection GoBoolExpressions
	for !context.Empty(current) && (false ||
		MatchDjotInlineNode(context, current, next, tokenizer.Paragraph, DocumentNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Emphasis, EmphasisNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Strong, StrongNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Highlighted, HighlightedNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Subscript, SubscriptNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Superscript, SuperscriptNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Insert, InsertNode) ||
		MatchDjotInlineNode(context, current, next, tokenizer.Delete, DeleteNode) ||
		MatchDjotTextNode(context, current, next, tokenizer.Autolink, AutolinkNode) ||
		MatchDjotTextNode(context, current, next, tokenizer.Symbols, SymbolsNode) ||
		MatchDjotVerbatim(context, current, next) ||
		MatchDjotLink(context, current, next) ||
		MatchDjotEscapedSymbol(context, current, next) ||
		MatchDjotText(context, current, next)) {
		current = *next
		level = append(level, *context.Tree...)
	}
	*context.Tree = level
	return true
}

func DjotAst(text []byte) Tree[DjotNode] {
	tokens := tokenizer.DjotTokens(text)
	context := MatchContext{
		TokenReader: &TokenReader[tokenizer.DjotToken]{Text: text, Tokens: tokens},
		Tree:        &[]Tree[DjotNode]{},
	}
	current := 0
	if !MatchDjotAst(context, current, &current) {
		panic(fmt.Errorf("incorrect sequence of djot tokens"))
	}
	return (*context.Tree)[0]
}
