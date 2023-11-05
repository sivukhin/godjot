package djot_tokenizer

import (
	"bytes"
	"github.com/sivukhin/godjot/tokenizer"
	"strings"
)

func ProcessDjotInlineTokens(
	tokenStack *tokenizer.TokenStack[DjotToken],
	reader tokenizer.TextReader,
	state tokenizer.ReaderState,
) tokenizer.ReaderState {
inlineParsingLoop:
	for !reader.Empty(state) {
		openInline := tokenStack.LastLevel().FirstOrDefault()
		openInlineType := openInline.Type

		lastInline := tokenStack.LastLevel().LastOrDefault()

		// Check if verbatim is open - then we can't process any inline-level elements
		if openInlineType == VerbatimInline {
			next := MatchInlineToken(reader, state, VerbatimInline^tokenizer.Open)
			if !next.Matched() {
				state++
				continue
			}
			openToken := reader.Select(tokenizer.ReaderState(openInline.Start), tokenizer.ReaderState(openInline.End))
			closeToken := reader.Select(state, next)
			if strings.TrimLeft(openToken, "$") != closeToken {
				state = next
				continue
			}
			tokenStack.CloseLevelAt(tokenizer.Token[DjotToken]{
				Type:  VerbatimInline ^ tokenizer.Open,
				Start: int(state),
				End:   int(next),
			})
			state = next
			continue
		}

		// Try match inline attribute
		if attributes, next := MatchDjotAttribute(reader, state); next.Matched() {
			tokenStack.LastLevel().Push(tokenizer.Token[DjotToken]{
				Type:       Attribute,
				Start:      int(state),
				End:        int(next),
				Attributes: attributes,
			})
			state = next
			continue
		}

		// EscapedSymbolInline / SmartSymbolInline is non-paired tokens - so we should treat it separately
		for _, tokenType := range []DjotToken{EscapedSymbolInline, SmartSymbolInline} {
			if next := MatchInlineToken(reader, state, tokenType); next.Matched() {
				tokenStack.LastLevel().Push(tokenizer.Token[DjotToken]{Type: tokenType, Start: int(state), End: int(next)})
				state = next
				continue inlineParsingLoop
			}
		}

		for _, tokenType := range []DjotToken{
			RawFormatInline,
			VerbatimInline,
			ImageSpanInline,
			LinkUrlInline,
			LinkReferenceInline,
			AutolinkInline,
			EmphasisInline,
			StrongInline,
			HighlightedInline,
			SubscriptInline,
			SuperscriptInline,
			InsertInline,
			DeleteInline,
			FootnoteReferenceInline,
			SpanInline,
			SymbolsInline,
		} {
			// Closing tokens take precedence because of greedy nature of parsing
			next := MatchInlineToken(reader, state, tokenType^tokenizer.Open)
			// EmphasisInline / StrongInline elements must contain something in between of open and close tokens
			forbidClose := (tokenType == EmphasisInline && lastInline.Type == EmphasisInline && lastInline.End == int(state)) ||
				(tokenType == StrongInline && lastInline.Type == StrongInline && lastInline.End == int(state))
			if !forbidClose && next.Matched() && tokenStack.PopForgetUntil(tokenType) {
				tokenStack.CloseLevelAt(tokenizer.Token[DjotToken]{Type: tokenType ^ tokenizer.Open, Start: int(state), End: int(next)})
				state = next
				continue inlineParsingLoop
			}
			// RawFormatInline must be preceded by VerbatimInline inline element only closed properly
			if tokenType == RawFormatInline && lastInline.Type != VerbatimInline^tokenizer.Open {
				continue
			}
			// LinkReferenceInline / LinkUrlInline must be preceded by SpanInline / ImageSpanInline inline element only closed properly
			if (tokenType == LinkReferenceInline || tokenType == LinkUrlInline) &&
				lastInline.Type != SpanInline^tokenizer.Open && lastInline.Type != ImageSpanInline^tokenizer.Open {
				continue
			}
			next = MatchInlineToken(reader, state, tokenType)
			if next.Matched() {
				var attributes tokenizer.Attributes
				token := reader[state:next]
				if tokenType == VerbatimInline && bytes.HasPrefix(token, []byte("$`")) {
					attributes.Set(InlineMathKey, "")
				} else if tokenType == VerbatimInline && bytes.HasPrefix(token, []byte("$$`")) {
					attributes.Set(DisplayMathKey, "")
				}
				tokenStack.OpenLevelAt(tokenizer.Token[DjotToken]{
					Type:       tokenType,
					Start:      int(state),
					End:        int(next),
					Attributes: &attributes,
				})
				state = next
				continue inlineParsingLoop
			}
		}

		// No tokens matched - proceed with next symbol
		state++
	}
	return state
}

func BuildDjotTokens(document []byte) tokenizer.TokenList[DjotToken] {
	var (
		lineTokenizer = tokenizer.LineTokenizer{Document: document}

		blockLineOffset  = []int{0}
		blockTokenOffset = []int{0}

		blockTokens = []tokenizer.Token[DjotToken]{{Type: DocumentBlock, Start: 0, End: 0}}
		finalTokens = tokenizer.TokenList[DjotToken]{{Type: DocumentBlock, Start: 0, End: 0}}
		inlineStack = tokenizer.NewTokenStack[DjotToken]()
	)

	popMetadata := func() {
		blockLineOffset = blockLineOffset[:len(blockLineOffset)-1]
		blockTokenOffset = blockTokenOffset[:len(blockTokenOffset)-1]
		blockTokens = blockTokens[:len(blockTokens)-1]
	}
	openBlockLevel := func(token tokenizer.Token[DjotToken]) {
		if (token.Type == ParagraphBlock || token.Type == HeadingBlock) && inlineStack.Empty() {
			inlineStack.OpenLevelAt(token)
		}
		finalTokens.Push(token)
		blockTokenOffset = append(blockTokenOffset, len(finalTokens)-1)
		blockTokens = append(blockTokens, token)
	}
	closeBlockLevelsUntil := func(start, end, level int) {
		if !inlineStack.Empty() {
			inlineStack.PopForgetUntil(ParagraphBlock)
			inlineStack.CloseLevelAt(tokenizer.Token[DjotToken]{Type: ParagraphBlock, Start: start, End: start})
			for _, token := range inlineStack.Levels[0][1 : len(inlineStack.Levels[0])-1] {
				finalTokens.Push(token)
			}
			inlineStack = tokenizer.NewTokenStack[DjotToken]()
		}
		for i := len(blockTokens) - 1; i > level; i-- {
			finalTokens.Push(tokenizer.Token[DjotToken]{
				Type:  blockTokens[i].Type ^ tokenizer.Open,
				Start: start,
				End:   end,
			})
			delta := len(finalTokens) - 1 - blockTokenOffset[i]
			finalTokens[blockTokenOffset[i]].JumpToPair = delta
			finalTokens[len(finalTokens)-1].JumpToPair = -delta
			popMetadata()
		}
	}

	for {
		lineStart, lineEnd, eof := lineTokenizer.Scan()
		if eof {
			break
		}

		reader, state := tokenizer.TextReader(document[:lineEnd]), tokenizer.ReaderState(lineStart)
		lastBlock := blockTokens[len(blockTokens)-1]
		lastBlockType := lastBlock.Type

		// Try to match block element attribute ({...}) at the start of the line (only in case if last block token was [Document | Quote | ListItem | Div])
		if lastBlockType == DocumentBlock || lastBlockType == QuoteBlock || lastBlockType == ListItemBlock || lastBlockType == DivBlock {
			next := reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
			attributes, next := MatchDjotAttribute(reader, next)
			if next.Matched() {
				next = reader.EmptyOrWhiteSpace(next)
			}
			if next.Matched() {
				finalTokens.Push(tokenizer.Token[DjotToken]{
					Type:       Attribute,
					Start:      int(state),
					End:        int(next),
					Attributes: attributes,
				})
				continue
			}
		}

		// Skip optional padding for Heading & Quotes (#, > padding) and remember last matched block token
		resetBlockAt, potentialReset := 0, false
		for i := 0; i < len(blockTokens); i++ {
			blockToken := blockTokens[i]
			if blockToken.Type == ListItemBlock {
				state = reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
				if !reader.EmptyOrWhiteSpace(state).Matched() && int(state)-lineStart <= blockLineOffset[i] {
					potentialReset = true
					break
				}
				resetBlockAt = i
			} else if blockToken.Type == QuoteBlock || blockToken.Type == HeadingBlock {
				_, next := MatchBlockToken(reader, state, blockToken.Type)
				if !next.Matched() {
					potentialReset = true
					break
				}
				state = next
				resetBlockAt = i
			}
		}

		if int(state) != lineStart {
			finalTokens.Push(tokenizer.Token[DjotToken]{Type: Padding, Start: lineStart, End: int(state)})
		}

		// Check for empty line and collapse all levels until resetBlockAt
		if (lastBlockType != CodeBlock || potentialReset) && reader.EmptyOrWhiteSpace(state).Matched() {
			closeBlockLevelsUntil(int(state), int(state), resetBlockAt)
			continue
		}

		// Check if last block is CodeBlock - then any block level logic should be disabled until we close this block
		if lastBlockType == CodeBlock {
			token, next := MatchBlockToken(reader, state, CodeBlock)
			if next.Matched() && lastBlock.PrefixLength(document, '`') <= token.PrefixLength(document, '`') && token.Attributes.Size() == 0 {
				closeBlockLevelsUntil(token.Start, token.End, len(blockTokens)-2)
			}
			continue
		}

		// Check if we can close DivBlock
		if lastBlockType == DivBlock {
			token, next := MatchBlockToken(reader, state, DivBlock)
			if next.Matched() && lastBlock.Length() <= token.Length() && token.Attributes.Size() == 0 {
				closeBlockLevelsUntil(token.Start, token.End, len(blockTokens)-2)
				continue
			}
		}

		// Main loop for matching of block level signs at the start of the line
	blockParsingLoop:
		for {
			lastBlock = blockTokens[len(blockTokens)-1]
			lastBlockType = lastBlock.Type

			// Check if thematic break finishes the line
			if thematicBreak, next := MatchBlockToken(reader, state, ThematicBreakToken); next.Matched() {
				finalTokens = append(finalTokens, tokenizer.Token[DjotToken]{
					Type:  ThematicBreakToken,
					Start: thematicBreak.Start,
					End:   thematicBreak.End,
				})
				continue blockParsingLoop
			}

			// Calculate potential reset level for list items due to indentation
			state = reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
			resetListPosition := -1
			for i := len(blockTokens) - 1; i >= 0; i-- {
				blockToken := blockTokens[i]
				if blockToken.Type == ListItemBlock && blockLineOffset[i] >= int(state)-lineStart {
					resetListPosition = i
				}
			}

			// Heading & CodeBlock can't have nested block level content
			// Paragraph too - but there are subtle rules for list item handling, so we can't break for paragraphs here
			if listItem, next := MatchBlockToken(reader, state, ListItemBlock); next.Matched() && lastBlockType != HeadingBlock && lastBlockType != CodeBlock {
				if resetListPosition != -1 {
					closeBlockLevelsUntil(int(state), int(state), resetListPosition-1)
				}
				// If we found list item which fits some previously defined hierarchy - then we will add it unconditionally
				if resetListPosition != -1 || lastBlockType != ParagraphBlock && lastBlockType != HeadingBlock && lastBlockType != CodeBlock {
					openBlockLevel(tokenizer.Token[DjotToken]{Type: ListItemBlock, Start: listItem.Start, End: listItem.End})
					blockLineOffset = append(blockLineOffset, listItem.Start-lineStart)
					state = next
					continue blockParsingLoop
				}
			}

			if lastBlockType == ParagraphBlock || lastBlockType == HeadingBlock {
				state = ProcessDjotInlineTokens(&inlineStack, reader, state)
				break blockParsingLoop
			}
			if lastBlockType == CodeBlock {
				break blockParsingLoop
			}

			if resetListPosition != -1 {
				closeBlockLevelsUntil(int(state), int(state), resetListPosition-1)
				continue blockParsingLoop
			}

			// Handle all other block elements - ParagraphBlock must be last item in the sequence
			for _, tokenType := range []DjotToken{
				HeadingBlock,
				QuoteBlock,
				ListItemBlock,
				CodeBlock,
				DivBlock,
				PipeTableBlock,
				ReferenceDefBlock,
				FootnoteDefBlock,
				ParagraphBlock,
			} {
				block, next := MatchBlockToken(reader, state, tokenType)
				if !next.Matched() {
					continue
				}
				// Forbid nesting FootnoteDefBlock and ReferenceDefBlock - they should be only on top level of the document
				if (tokenType == FootnoteDefBlock || tokenType == ReferenceDefBlock) && lastBlockType != DocumentBlock {
					continue
				}
				openBlockLevel(block)
				blockLineOffset = append(blockLineOffset, block.Start-lineStart)
				state = next
				continue blockParsingLoop
			}

			break
		}
	}

	closeBlockLevelsUntil(len(document), len(document), -1)
	return finalTokens
}
