package parser

import (
	"fmt"
	"github.com/sivukhin/godjot/tokenizer"
)

type DjotNode int

const (
	TextNode = iota
	DocumentNode
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
	LineBreakNode
)

func (node DjotNode) String() string {
	switch node {
	case TextNode:
		return "TextNode"
	case DocumentNode:
		return "DocumentNode"
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
	case AutolinkNode:
		return "AutolinkNode"
	case SymbolsNode:
		return "SymbolsNode"
	case VerbatimNode:
		return "VerbatimNode"
	case LineBreakNode:
		return "LineBreakNode"
	default:
		panic(fmt.Errorf("unexpected djot node type: %d", node))
	}
}

func MatchDjotSimpleNode(
	reader *TokenReader[tokenizer.DjotToken],
	token tokenizer.DjotToken,
	node DjotNode,
	tree *[]Tree[DjotNode],
) bool {
	if _, match := reader.MatchToken(token, reader); !match {
		return false
	}
	if !MatchDjotAst(reader, tree) {
		return false
	}
	if _, match := reader.MatchToken(token|tokenizer.Close, reader); !match {
		return false
	}
	*tree = []Tree[DjotNode]{{Type: node, Children: *tree}}
	return true
}

func MatchDjotEscapedSymbol(reader *TokenReader[tokenizer.DjotToken], tree *[]Tree[DjotNode]) bool {
	token, ok := reader.MatchToken(tokenizer.Escaped, reader)
	if !ok {
		return false
	}
	if reader.Text[token.End-1] == '\n' {
		*tree = []Tree[DjotNode]{{Type: LineBreakNode}}
	} else {
		*tree = []Tree[DjotNode]{{Type: TextNode, Text: string(reader.Text[token.Start+1 : token.End])}}
	}
	return true
}

func MatchDjotText(reader *TokenReader[tokenizer.DjotToken], tree *[]Tree[DjotNode]) bool {
	token, ok := reader.MatchToken(tokenizer.Raw, reader)
	if !ok {
		return false
	}
	*tree = []Tree[DjotNode]{{Type: TextNode, Text: string(reader.Text[token.Start:token.End])}}
	return true
}

func MatchDjotAst(reader *TokenReader[tokenizer.DjotToken], tree *[]Tree[DjotNode]) bool {
	level := make([]Tree[DjotNode], 0)
	//goland:noinspection GoBoolExpressions
	for !reader.Empty() && (false ||
		MatchDjotSimpleNode(reader, tokenizer.Doc, DocumentNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Emphasis, EmphasisNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Strong, StrongNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Highlighted, HighlightedNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Subscript, SubscriptNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Superscript, SuperscriptNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Insert, InsertNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Delete, DeleteNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Autolink, AutolinkNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Symbols, SymbolsNode, tree) ||
		MatchDjotSimpleNode(reader, tokenizer.Verbatim, VerbatimNode, tree) ||
		MatchDjotEscapedSymbol(reader, tree) ||
		MatchDjotText(reader, tree)) {
		level = append(level, *tree...)
	}
	*tree = level
	return true
}

func DjotAst(text []byte) Tree[DjotNode] {
	tokens := tokenizer.DjotTokens(text)
	reader := TokenReader[tokenizer.DjotToken]{Text: text, Tokens: tokens}
	level := make([]Tree[DjotNode], 0)
	if !MatchDjotAst(&reader, &level) {
		panic(fmt.Errorf("incorrect sequence of djot tokens"))
	}
	return level[0]
}
