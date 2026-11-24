package djot_tokenizer

import (
	"github.com/sivukhin/godjot/v2/tokenizer"
)

var (
	DollarByteMask                 = tokenizer.NewByteMask([]byte("$"))
	BacktickByteMask               = tokenizer.NewByteMask([]byte("`"))
	SmartSymbolByteMask            = tokenizer.NewByteMask([]byte("\n'\""))
	AlphaNumericSymbolByteMask     = tokenizer.NewByteMask([]byte("+-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	AsciiPunctuationSymbolByteMask = tokenizer.NewByteMask([]byte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"))
	InlineTokenStartSymbol         = tokenizer.NewByteMask([]byte("!\"$'()*+-.:<=>[\\]^_`{|}~")).Or(tokenizer.SpaceNewLineByteMask)
)

const RecordStartSymbol = true

type Tokenizer struct {
	StartSymbols map[byte]struct{}
}

func New() Tokenizer {
	return Tokenizer{make(map[byte]struct{})}
}

func (t Tokenizer) MatchInlineToken(
	r tokenizer.TextReader,
	s tokenizer.ReaderState,
	tokenType DjotToken,
) (tokenizer.ReaderState, bool) {
	state, ok := matchInlineToken(r, s, tokenType)
	//goland:noinspection GoBoolExpressions
	if RecordStartSymbol && ok && !r.IsEmpty(s) {
		t.StartSymbols[r[s]] = struct{}{}
	}
	return state, ok
}

func matchInlineToken(
	r tokenizer.TextReader,
	s tokenizer.ReaderState,
	tokenType DjotToken,
) (tokenizer.ReaderState, bool) {
	fail := func() (tokenizer.ReaderState, bool) { return 0, false }

	switch tokenType {
	case RawFormatInline:
		return r.Token2(s, [...]byte{'{', '='})
	case RawFormatInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{'}'})

	case VerbatimInline:
		next, ok := r.MaskRepeat(s, DollarByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
		// $ - inline math, $$ - display math, more dollars - unknown syntax
		if next-s > 2 {
			return fail()
		}
		return r.MaskRepeat(next, BacktickByteMask, 1)
	case VerbatimInline ^ tokenizer.Open:
		return r.MaskRepeat(s, BacktickByteMask, 1)
	case ImageSpanInline:
		return r.Token2(s, [...]byte{'!', '['})
	case SpanInline:
		return r.Token1(s, [...]byte{'['})
	case SpanInline ^ tokenizer.Open, ImageSpanInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{']'})
	case LinkUrlInline:
		return r.Token1(s, [...]byte{'('})
	case LinkUrlInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{')'})
	case LinkReferenceInline:
		return r.Token1(s, [...]byte{'['})
	case LinkReferenceInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{']'})
	case AutolinkInline:
		return r.Token1(s, [...]byte{'<'})
	case AutolinkInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{'>'})
	case EscapedSymbolInline:
		next, ok := r.Token1(s, [...]byte{'\\'})
		if !ok {
			return fail()
		}
		if r.IsEmpty(next) {
			return fail()
		}
		if asciiNext, ok := r.Mask(next, AsciiPunctuationSymbolByteMask); ok {
			return asciiNext, ok
		}
		next, ok = r.MaskRepeat(next, tokenizer.SpaceByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
		return r.Token1(next, [...]byte{'\n'})
	case EmphasisInline:
		if next, ok := r.Token2(s, [...]byte{'{', '_'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'_'}); ok && !r.HasMask(next, tokenizer.SpaceNewLineByteMask) && !r.HasToken1(next, [...]byte{'}'}) {
			return next, ok
		}
		return fail()
	case EmphasisInline ^ tokenizer.Open:
		if next, ok := r.Token2(s, [...]byte{'_', '}'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'_'}); ok && s > 0 && !r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case StrongInline:
		if next, ok := r.Token2(s, [...]byte{'{', '*'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'*'}); ok && !r.HasMask(next, tokenizer.SpaceNewLineByteMask) && !r.HasToken1(next, [...]byte{'}'}) {
			return next, ok
		}
		return fail()
	case StrongInline ^ tokenizer.Open:
		if next, ok := r.Token2(s, [...]byte{'*', '}'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'*'}); ok && s > 0 && !r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case HighlightedInline:
		return r.Token2(s, [...]byte{'{', '='})
	case HighlightedInline ^ tokenizer.Open:
		return r.Token2(s, [...]byte{'=', '}'})
	case SubscriptInline:
		if next, ok := r.Token2(s, [...]byte{'{', '~'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'~'}); ok && !r.HasMask(next, tokenizer.SpaceNewLineByteMask) && !r.HasToken1(next, [...]byte{'}'}) {
			return next, ok
		}
		return fail()
	case SubscriptInline ^ tokenizer.Open:
		if next, ok := r.Token2(s, [...]byte{'~', '}'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'~'}); ok && s > 0 && !r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case SuperscriptInline:
		if next, ok := r.Token2(s, [...]byte{'{', '^'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'^'}); ok && !r.HasMask(next, tokenizer.SpaceNewLineByteMask) && !r.HasToken1(next, [...]byte{'}'}) {
			return next, ok
		}
		return fail()
	case SuperscriptInline ^ tokenizer.Open:
		if next, ok := r.Token2(s, [...]byte{'^', '}'}); ok {
			return next, ok
		}
		if next, ok := r.Token1(s, [...]byte{'^'}); ok && s > 0 && !r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case InsertInline:
		return r.Token2(s, [...]byte{'{', '+'})
	case InsertInline ^ tokenizer.Open:
		return r.Token2(s, [...]byte{'+', '}'})
	case DeleteInline:
		return r.Token2(s, [...]byte{'{', '-'})
	case DeleteInline ^ tokenizer.Open:
		return r.Token2(s, [...]byte{'-', '}'})
	case FootnoteReferenceInline:
		return r.Token2(s, [...]byte{'[', '^'})
	case FootnoteReferenceInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{']'})
	case SymbolsInline:
		next, ok := r.Token1(s, [...]byte{':'})
		if !ok {
			return fail()
		}
		if word, ok := r.MaskRepeat(next, AlphaNumericSymbolByteMask, 0); ok && r.HasToken1(word, [...]byte{':'}) {
			return next, true
		}
		return fail()
	case SymbolsInline ^ tokenizer.Open:
		return r.Token1(s, [...]byte{':'})
	case PipeTableSeparator:
		next, ok := r.Token1(s, [...]byte{'|'})
		if !ok {
			return fail()
		}
		return r.MaskRepeat(next, tokenizer.SpaceByteMask, 0)
	case PipeTableSeparator ^ tokenizer.Open:
		s, ok := r.MaskRepeat(s, tokenizer.SpaceByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")

		if next, ok := r.Token1(s, [...]byte{'|'}); ok {
			if r.IsEmptyOrWhiteSpace(next) {
				return next, true
			}
			return s, true
		}
		return fail()
	case SmartSymbolInline:
		if next, ok := r.Token1(s, [...]byte{'{'}); ok {
			return r.Mask(next, SmartSymbolByteMask)
		}
		if next, ok := r.Mask(s, SmartSymbolByteMask); ok {
			if r.HasToken1(next, [...]byte{'}'}) {
				return next + 1, true
			}
			return next, true
		}
		if next, ok := r.Token3(s, [...]byte{'.', '.', '.'}); ok {
			return next, ok
		}
		if next, ok := r.ByteRepeat(s, '-', 2); ok {
			return next, ok
		}
		return fail()
	default:
		tokenizer.Assertf(false, "unexpected djot inline token type: %v(%d)", tokenType, tokenType)
		return fail()
	}
}
