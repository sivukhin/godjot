package tokenizer

import (
	"strings"
)

func ProcessDjotInlineTokens(
	tokenizer *Tokenizer[DjotToken],
	reader TextReader,
	state ReaderState,
) ReaderState {
inlineParsingLoop:
	for !reader.Empty(state) {
		openInline := tokenizer.LastLevel().First()
		openInlineType := openInline.Type

		lastInline := tokenizer.LastLevel().Last()

		// Check if verbatim is open - then we can't process any inline-level elements
		if openInlineType == Verbatim {
			next := MatchInlineToken(reader, state, Verbatim^Open)
			openToken := reader.Select(ReaderState(openInline.Start), ReaderState(openInline.End))
			closeToken := reader.Select(state, next)
			if next.Matched() && strings.TrimLeft(openToken, "$") == closeToken {
				tokenizer.CloseLevelAt(Verbatim^Open, int(state), int(next))
				state = next
			} else {
				state++
			}
			continue
		}

		// Try match inline attribute
		if attributes, next := MatchDjotAttribute(reader, state); next.Matched() {
			tokenizer.AddLengthToken(Attribute, int(state), int(next), attributes)
			state = next
			continue
		}

		// Escaped is non-paired token - so we should treat it separately
		if next := MatchInlineToken(reader, state, Escaped); next.Matched() {
			tokenizer.AddLengthToken(Escaped, int(state), int(next))
			state = next
			continue
		}

		for _, tokenType := range []DjotToken{
			Verbatim,
			Span,
			LinkUrl,
			LinkReference,
			Autolink,
			Emphasis,
			Strong,
			Highlighted,
			Subscript,
			Superscript,
			Insert,
			Delete,
			FootnoteReference,
			Symbols,
			RawFormat,
		} {
			// Closing tokens take precedence because of greedy nature of parsing
			next := MatchInlineToken(reader, state, tokenType^Open)
			// Emphasis / Strong elements must contain something in between of open and close tokens
			forbidClose := (tokenType == Emphasis && lastInline.Type == Emphasis && lastInline.End == int(state)) ||
				(tokenType == Strong && lastInline.Type == Strong && lastInline.End == int(state))
			if !forbidClose && next.Matched() && tokenizer.PopForgetUntil(tokenType) {
				tokenizer.CloseLevelAt(tokenType^Open, int(state), int(next))
				state = next
				continue inlineParsingLoop
			}
			// RawFormat must be preceded by Verbatim inline element only closed properly
			if tokenType == RawFormat && lastInline.Type != Verbatim^Open {
				continue
			}
			// LinkReference / LinkUrl must be preceded by Span inline element only closed properly
			if (tokenType == LinkReference || tokenType == LinkUrl) && lastInline.Type != Span^Open {
				continue
			}
			next = MatchInlineToken(reader, state, tokenType)
			if next.Matched() {
				tokenizer.OpenLevelAt(tokenType, int(state), int(next))
				state = next
				continue inlineParsingLoop
			}
		}

		// No tokens matched - proceed with next symbol
		state++
	}
	return state
}

func DjotTokens(document []byte) []Token[DjotToken] {
	var (
		lineTokenizer = LineTokenizer{Document: document}

		paddingResetLevel = []int{0}
		blockLineOffset   = []int{0}
		blockTokenOffset  = []int{0}
		blockTokens       = []Token[DjotToken]{{Type: DocumentBlock, Start: 0, End: 0}}
		finalTokens       = []Token[DjotToken]{{Type: DocumentBlock, Start: 0, End: 0}}

		inlineTokenizer = NewTokenizer[DjotToken]()
	)

	popMetadata := func() {
		paddingResetLevel = paddingResetLevel[:len(paddingResetLevel)-1]
		blockLineOffset = blockLineOffset[:len(blockLineOffset)-1]
		blockTokenOffset = blockTokenOffset[:len(blockTokenOffset)-1]
		blockTokens = blockTokens[:len(blockTokens)-1]
	}
	openBlockLevel := func(token Token[DjotToken]) {
		if token.Type == ParagraphToken && len(inlineTokenizer.Levels) == 0 {
			inlineTokenizer.OpenLevelAt(ParagraphToken, 0, 0)
		}
		blockTokenOffset = append(blockTokenOffset, len(finalTokens))
		blockTokens = append(blockTokens, token)
		finalTokens = append(finalTokens, token)
	}
	closeBlockLevelsUntil := func(start, end, level int) {
		if len(inlineTokenizer.Levels) > 0 {
			inlineTokenizer.PopForgetUntil(ParagraphToken)
			finalTokens = append(finalTokens, inlineTokenizer.Levels[0][1:len(inlineTokenizer.Levels[0])]...)
			inlineTokenizer = NewTokenizer[DjotToken]()
		}
		for i := len(blockTokens) - 1; i > level; i-- {
			delta := len(finalTokens) - blockTokenOffset[i]
			finalTokens = append(finalTokens, Token[DjotToken]{
				Type:  blockTokens[i].Type ^ Open,
				Start: start,
				End:   end,
			})
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

		reader, state := TextReader(document[:lineEnd]), ReaderState(lineStart)
		lastBlock := blockTokens[len(blockTokens)-1]
		lastBlockType := lastBlock.Type

		// Try to match block element attribute ({...}) at the start of the line (only in case if last block token was [Document | Quote | ListItem | Div])
		if lastBlockType == DocumentBlock || lastBlockType == QuoteBlock || lastBlockType == ListItemBlock || lastBlockType == DivBlock {
			next := reader.MaskRepeat(state, SpaceByteMask, 0)
			attributes, next := MatchDjotAttribute(reader, next)
			if next.Matched() {
				next = reader.EmptyOrWhiteSpace(next)
			}
			if next.Matched() {
				finalTokens = append(finalTokens, Token[DjotToken]{
					Type:       Attribute,
					Start:      int(state),
					End:        int(next),
					Attributes: attributes,
				})
				continue
			}
		}

		// Skip optional padding for Heading & Quotes (#, > padding) and remember last matched block token
		lastMatchedPadding := 0
		for i := paddingResetLevel[len(paddingResetLevel)-1]; i < len(blockTokens); i++ {
			blockToken := blockTokens[i]
			if blockToken.Type != QuoteBlock && blockToken.Type != HeadingBlock {
				continue
			}
			_, next := MatchBlockToken(reader, state, blockToken.Type)
			if !next.Matched() {
				continue
			}
			state = next
			lastMatchedPadding = i
		}

		// Check for empty line and collapse all levels until lastMatchedPadding
		if lastBlockType != CodeBlock && reader.EmptyOrWhiteSpace(state).Matched() {
			closeBlockLevelsUntil(int(state), int(state), lastMatchedPadding)
			continue
		}

		// Check if thematic break finishes the line
		if thematicBreak, next := MatchBlockToken(reader, state, ThematicBreakToken); lastBlockType != CodeBlock && next.Matched() {
			finalTokens = append(finalTokens, Token[DjotToken]{Type: ThematicBreakToken, Start: thematicBreak.Start, End: thematicBreak.End})
			closeBlockLevelsUntil(thematicBreak.End, thematicBreak.End, lastMatchedPadding)
			continue
		}

		// Check if last block is CodeBlock - then any block level logic should be disabled until we close this block
		if lastBlockType == CodeBlock {
			token, next := MatchBlockToken(reader, state, CodeBlock)
			if next.Matched() && lastBlock.Length() <= token.Length() && len(token.Attributes) == 0 {
				closeBlockLevelsUntil(token.Start, token.End, len(blockTokens)-2)
			}
			continue
		}

		// Check if we can close DivBlock
		if lastBlockType == DivBlock {
			token, next := MatchBlockToken(reader, state, DivBlock)
			if next.Matched() && lastBlock.Length() <= token.Length() && len(token.Attributes) == 0 {
				closeBlockLevelsUntil(token.Start, token.End, len(blockTokens)-2)
				continue
			}
		}

		// Main loop for matching of block level signs at the start of the line
	blockParsingLoop:
		for {
			lastBlock = blockTokens[len(blockTokens)-1]
			lastBlockType = lastBlock.Type

			// Heading & CodeBlock can't have nested block level content
			// Paragraph too - but there are subtle rules for list item handling, so we can't break for paragraphs here
			if lastBlockType == HeadingBlock || lastBlockType == CodeBlock {
				break blockParsingLoop
			}

			if listItem, next := MatchBlockToken(reader, state, ListItemBlock); next.Matched() {
				resetListPosition := -1
				for i := len(blockTokens) - 1; i >= paddingResetLevel[len(paddingResetLevel)-1]; i-- {
					blockToken := blockTokens[i]
					if blockToken.Type == ListItemBlock && blockLineOffset[i] >= listItem.Start-lineStart {
						resetListPosition = i
					}
				}
				// If we found list item which fits some previously defined hierarchy - then we will add it unconditionally
				if resetListPosition != -1 {
					closeBlockLevelsUntil(listItem.Start, listItem.Start, resetListPosition-1)
				}
				if resetListPosition != -1 || lastBlockType != ParagraphToken && lastBlockType != HeadingBlock && lastBlockType != CodeBlock {
					openBlockLevel(Token[DjotToken]{Type: ListItemBlock, Start: listItem.Start, End: listItem.End})
					paddingResetLevel = append(paddingResetLevel, paddingResetLevel[len(paddingResetLevel)-1])
					blockLineOffset = append(blockLineOffset, listItem.Start-lineStart)
					state = next
					continue blockParsingLoop
				}
			}

			if lastBlockType == ParagraphToken {
				state = ProcessDjotInlineTokens(&inlineTokenizer, reader, state)
				break blockParsingLoop
			}

			// Handle all other block elements - ParagraphToken must be last item in the sequence
			for _, tokenType := range []DjotToken{
				HeadingBlock,
				QuoteBlock,
				ListItemBlock,
				CodeBlock,
				DivBlock,
				PipeTableBlock,
				ReferenceDefBlock,
				FootnoteDefBlock,
				ParagraphToken,
			} {
				block, next := MatchBlockToken(reader, state, tokenType)
				if !next.Matched() {
					continue
				}
				// Forbid nesting FootnoteDefBlock and ReferenceDefBlock - they should be only on top level of the document
				if (tokenType == FootnoteDefBlock || tokenType == ReferenceDefBlock) && lastBlockType == DocumentBlock {
					continue
				}
				openBlockLevel(block)
				blockLineOffset = append(blockLineOffset, block.Start-lineStart)
				if tokenType == DivBlock || tokenType == CodeBlock {
					paddingResetLevel = append(paddingResetLevel, len(paddingResetLevel))
				} else {
					paddingResetLevel = append(paddingResetLevel, paddingResetLevel[len(paddingResetLevel)-1])
				}
				state = next
				continue blockParsingLoop
			}

			break
		}
	}

	closeBlockLevelsUntil(len(document), len(document), -1)
	return finalTokens
}
