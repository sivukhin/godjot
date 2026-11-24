package djot_tokenizer

import (
	"bytes"
	"strings"

	"github.com/sivukhin/godjot/v2/tokenizer"
)

func BuildInlineDjotTokens(
	document []byte,
	parts ...tokenizer.Range,
) []tokenizer.Token[DjotToken] {
	tok := New()

	if len(parts) == 0 {
		parts = []tokenizer.Range{{Start: 0, End: len(document)}}
	}

	tokenStack := tokenizer.NewTokenStack[DjotToken]()
	leftDocumentPosition, rightDocumentPosition := parts[0].Start, parts[len(parts)-1].End
	tokenStack.OpenLevelAt(tokenizer.Token[DjotToken]{
		Type:  ParagraphBlock,
		Start: leftDocumentPosition,
		End:   leftDocumentPosition,
	})
	for _, part := range parts {
		reader, state := tokenizer.TextReader(document[:part.End]), part.Start
		tokenStack.LastLevel().FillUntil(part.Start, Ignore)

	inlineParsingLoop:
		for !reader.IsEmpty(state) {
			openInline := tokenStack.LastLevel().FirstOrDefault()
			openInlineType := openInline.Type

			lastInline := tokenStack.LastLevel().LastOrDefault()

			// Check if verbatim is open - then we can't process any inline-level elements
			if openInlineType == VerbatimInline {
				next, ok := tok.MatchInlineToken(reader, state, VerbatimInline^tokenizer.Open)
				if !ok {
					state++
					continue
				}
				openToken := reader.Select(openInline.Start, openInline.End)
				closeToken := reader.Select(state, next)
				if strings.TrimLeft(openToken, "$") != closeToken {
					state = next
					continue
				}
				tokenStack.CloseLevelAt(tokenizer.Token[DjotToken]{
					Type:  VerbatimInline ^ tokenizer.Open,
					Start: state,
					End:   next,
				})
				state = next
				continue
			}

			// Try match inline attribute
			if attributes, next, ok := MatchDjotAttribute(reader, state); ok {
				tokenStack.LastLevel().Push(tokenizer.Token[DjotToken]{
					Type:       Attribute,
					Start:      state,
					End:        next,
					Attributes: attributes,
				})
				state = next
				continue
			}

			if !InlineTokenStartSymbol.Has(reader[state]) {
				state++
				continue
			}

			// EscapedSymbolInline / SmartSymbolInline is non-paired tokens - so we should treat it separately
			for _, tokenType := range []DjotToken{EscapedSymbolInline, SmartSymbolInline} {
				if next, ok := tok.MatchInlineToken(reader, state, tokenType); ok {
					tokenStack.LastLevel().Push(tokenizer.Token[DjotToken]{Type: tokenType, Start: state, End: next})
					state = next
					continue inlineParsingLoop
				}
			}

			for _, tokenType := range [...]DjotToken{
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
				PipeTableSeparator,
			} {
				// Closing tokens take precedence because of greedy nature of parsing
				next, ok := tok.MatchInlineToken(reader, state, tokenType^tokenizer.Open)
				// EmphasisInline / StrongInline elements must contain something in between of open and close tokens
				forbidClose := (tokenType == EmphasisInline && lastInline.Type == EmphasisInline && lastInline.End == state) ||
					(tokenType == StrongInline && lastInline.Type == StrongInline && lastInline.End == state)
				if !forbidClose && ok && tokenStack.PopForgetUntil(tokenType) {
					tokenStack.CloseLevelAt(tokenizer.Token[DjotToken]{Type: tokenType ^ tokenizer.Open, Start: state, End: next})
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
				next, ok = tok.MatchInlineToken(reader, state, tokenType)
				if ok {
					var attributes tokenizer.Attributes
					token := reader[state:next]
					if tokenType == VerbatimInline && bytes.HasPrefix(token, []byte("$$")) {
						attributes.Set(DisplayMathKey, "")
					} else if tokenType == VerbatimInline && bytes.HasPrefix(token, []byte("$")) {
						attributes.Set(InlineMathKey, "")
					}
					tokenStack.OpenLevelAt(tokenizer.Token[DjotToken]{
						Type:       tokenType,
						Start:      state,
						End:        next,
						Attributes: attributes,
					})
					state = next
					continue inlineParsingLoop
				}
			}

			// No tokens matched - proceed with next symbol
			state++
		}
	}
	if tokenStack.LastLevel().FirstOrDefault().Type == VerbatimInline {
		tokenStack.CloseLevelAt(tokenizer.Token[DjotToken]{Type: VerbatimInline ^ tokenizer.Open, Start: rightDocumentPosition, End: rightDocumentPosition})
	}
	tokenStack.PopForgetUntil(ParagraphBlock)
	tokenStack.CloseLevelAt(tokenizer.Token[DjotToken]{Type: ParagraphBlock, Start: rightDocumentPosition, End: rightDocumentPosition})
	tokens := *tokenStack.LastLevel()
	return tokens[1 : len(tokens)-1]
}

func BuildDjotTokens(document []byte) tokenizer.TokenList[DjotToken] {
	var (
		lineTokenizer = tokenizer.LineTokenizer{Document: document}

		inlineParts = tokenizer.Ranges{}

		blockLineOffset  = []int{0}
		blockTokenOffset = []int{0}

		blockTokens = []tokenizer.Token[DjotToken]{{Type: DocumentBlock, Start: 0, End: 0}}
		finalTokens = []tokenizer.Token[DjotToken]{{Type: DocumentBlock, Start: 0, End: 0}}
	)

	popMetadata := func() {
		blockLineOffset = blockLineOffset[:len(blockLineOffset)-1]
		blockTokenOffset = blockTokenOffset[:len(blockTokenOffset)-1]
		blockTokens = blockTokens[:len(blockTokens)-1]
	}
	openBlockLevel := func(token tokenizer.Token[DjotToken]) {
		finalTokens = append(finalTokens, token)
		blockTokenOffset = append(blockTokenOffset, len(finalTokens)-1)
		blockTokens = append(blockTokens, token)
	}
	closeBlockLevelsUntil := func(start, end, level int) {
		if len(inlineParts) != 0 && blockTokens[len(blockTokens)-1].Type == CodeBlock {
			for _, inlinePart := range inlineParts {
				finalTokens = append(finalTokens, tokenizer.Token[DjotToken]{Start: inlinePart.Start, End: inlinePart.End})
			}
			inlineParts = nil
		} else if len(inlineParts) != 0 {
			finalTokens = append(finalTokens, BuildInlineDjotTokens(document, inlineParts...)...)
			inlineParts = nil
		}
		for i := len(blockTokens) - 1; i > level; i-- {
			finalTokens = append(finalTokens, tokenizer.Token[DjotToken]{
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

		reader, state := tokenizer.TextReader(document[:lineEnd]), lineStart
		lastBlock := blockTokens[len(blockTokens)-1]
		lastBlockType := lastBlock.Type

		// Try to match block element attribute ({...}) at the start of the line (only in case if last block token was [Document | Quote | ListItem | Div])
		if lastBlockType == DocumentBlock || lastBlockType == QuoteBlock || lastBlockType == ListItemBlock || lastBlockType == DivBlock {
			next, ok := reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
			tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

			attributes, next, ok := MatchDjotAttribute(reader, next)
			if ok {
				next, ok = reader.EmptyOrWhiteSpace(next)
			}
			if ok {
				finalTokens = append(finalTokens, tokenizer.Token[DjotToken]{
					Type:       Attribute,
					Start:      state,
					End:        next,
					Attributes: attributes,
				})
				continue
			}
		}

		lastDivAt := -1
		for i := 0; i < len(blockTokens); i++ {
			blockToken := blockTokens[i]
			if blockToken.Type == DivBlock {
				lastDivAt = i
			}
		}
		// Skip optional padding for Heading & Quotes (#, > padding) and remember last matched block token
		resetBlockAt, potentialReset := 0, false
		for i := 0; i < len(blockTokens); i++ {
			blockToken := blockTokens[i]
			if blockToken.Type == ListItemBlock || blockToken.Type == FootnoteDefBlock {
				next, ok := reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
				tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

				if !reader.IsEmptyOrWhiteSpace(next) && next-lineStart <= blockLineOffset[i] {
					potentialReset = true
					break
				}
				resetBlockAt = i
			} else if blockToken.Type == ReferenceDefBlock {
				next, ok := reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
				tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
				if next-lineStart <= blockLineOffset[i] {
					potentialReset = true
					break
				}
				resetBlockAt = i
			} else if blockToken.Type == QuoteBlock || blockToken.Type == HeadingBlock {
				_, next, ok := MatchBlockToken(reader, state, blockToken.Type)
				if !ok {
					potentialReset = true
					break
				}
				state = next
				resetBlockAt = i
			} else if blockToken.Type != ParagraphBlock && blockToken.Type != HeadingBlock && blockToken.Type != PipeTableCaptionBlock {
				resetBlockAt = i
			}
		}

		if (lastBlockType != CodeBlock || potentialReset) && reader.IsEmptyOrWhiteSpace(state) {
			closeBlockLevelsUntil(state, state, resetBlockAt)
			continue
		}

		// Force resets for ReferenceDefBlock if any
		if lastBlockType == ReferenceDefBlock {
			closeBlockLevelsUntil(state, state, resetBlockAt)
		}

		// Check if last block is CodeBlock - then any block level logic should be disabled until we close this block
		if lastBlockType == CodeBlock {
			token, _, ok := MatchBlockToken(reader, state, CodeBlock)
			if ok && lastBlock.PrefixLength(document, '`') <= token.PrefixLength(document, '`') && token.Attributes.Size() == 0 {
				closeBlockLevelsUntil(token.Start, token.End, len(blockTokens)-2)
			} else {
				inlineParts = append(inlineParts, tokenizer.Range{Start: state, End: lineEnd})
			}
			continue
		}

		// Check if we can close DivBlock
		if lastDivAt != -1 {
			token, _, ok := MatchBlockToken(reader, state, DivBlock)
			if ok && lastBlock.Length() <= token.Length() && token.Attributes.Size() == 0 {
				closeBlockLevelsUntil(token.Start, token.End, lastDivAt-1)
				continue
			}
		}

		// Main loop for matching of block level signs at the start of the line
	blockParsingLoop:
		for {
			lastBlock = blockTokens[len(blockTokens)-1]
			lastBlockType = lastBlock.Type

			// Check if thematic break finishes the line
			if thematicBreak, next, ok := MatchBlockToken(reader, state, ThematicBreakToken); ok {
				finalTokens = append(finalTokens, tokenizer.Token[DjotToken]{
					Type:  ThematicBreakToken,
					Start: thematicBreak.Start,
					End:   thematicBreak.End,
				})
				state = next
				continue blockParsingLoop
			}

			// Calculate potential reset level for list items due to indentation
			var ok bool
			state, ok = reader.MaskRepeat(state, tokenizer.SpaceByteMask, 0)
			tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

			resetListPosition := -1
			for i := len(blockTokens) - 1; i >= 0; i-- {
				blockToken := blockTokens[i]
				if blockToken.Type == ListItemBlock && blockLineOffset[i] >= state-lineStart {
					resetListPosition = i
				}
			}

			// Heading & CodeBlock can't have nested block level content
			// Paragraph too - but there are subtle rules for list item handling, so we can't break for paragraphs here
			if listItem, next, ok := MatchBlockToken(reader, state, ListItemBlock); ok && lastBlockType != HeadingBlock && lastBlockType != CodeBlock {
				if resetListPosition != -1 {
					closeBlockLevelsUntil(state, state, resetListPosition-1)
				}
				// If we found list item which fits some previously defined hierarchy - then we will add it unconditionally
				if resetListPosition != -1 || lastBlockType != ParagraphBlock && lastBlockType != HeadingBlock && lastBlockType != CodeBlock {
					openBlockLevel(tokenizer.Token[DjotToken]{Type: ListItemBlock, Start: listItem.Start, End: listItem.End})
					blockLineOffset = append(blockLineOffset, listItem.Start-lineStart)
					state = next
					continue blockParsingLoop
				}
			}

			if lastBlockType == ParagraphBlock || lastBlockType == HeadingBlock || lastBlockType == PipeTableCaptionBlock || lastBlockType == ReferenceDefBlock {
				inlineParts.Push(tokenizer.Range{Start: state, End: lineEnd})
				break blockParsingLoop
			}
			if lastBlockType == PipeTableBlock {
				inlineParts.Push(tokenizer.Range{Start: state, End: lineEnd})
				closeBlockLevelsUntil(lineEnd, lineEnd, len(blockTokens)-2)
				break blockParsingLoop
			}

			if lastBlockType == CodeBlock {
				break blockParsingLoop
			}

			if resetListPosition != -1 {
				closeBlockLevelsUntil(state, state, resetListPosition-1)
				continue blockParsingLoop
			}

			// Handle all other block elements - ParagraphBlock must be last item in the sequence
			for _, tokenType := range [...]DjotToken{
				FootnoteDefBlock,
				ReferenceDefBlock,
				HeadingBlock,
				QuoteBlock,
				ListItemBlock,
				CodeBlock,
				DivBlock,
				PipeTableBlock,
				PipeTableCaptionBlock,
				ParagraphBlock,
			} {
				// Allow FootnoteDefBlock and ReferenceDefBlock blocks only on top level of the document
				if lastBlockType != DocumentBlock && (tokenType == FootnoteDefBlock || tokenType == ReferenceDefBlock) {
					continue
				}
				block, next, ok := MatchBlockToken(reader, state, tokenType)
				if !ok {
					continue
				}
				if tokenType == PipeTableCaptionBlock && (len(finalTokens) == 0 || finalTokens[len(finalTokens)-1].Type != PipeTableBlock^tokenizer.Open) {
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
