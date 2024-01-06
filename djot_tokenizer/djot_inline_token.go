package djot_tokenizer

import (
	"github.com/sivukhin/godjot/tokenizer"
)

var (
	DollarByteMask                 = tokenizer.NewByteMask([]byte("$"))
	BacktickByteMask               = tokenizer.NewByteMask([]byte("`"))
	SmartSymbolByteMask            = tokenizer.NewByteMask([]byte("\n'\""))
	AlphaNumericSymbolByteMask     = tokenizer.NewByteMask([]byte("+-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	AsciiPunctuationSymbolByteMask = tokenizer.NewByteMask([]byte(" !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"))
)

func MatchInlineToken(
	r tokenizer.TextReader,
	s tokenizer.ReaderState,
	tokenType DjotToken,
) (tokenizer.ReaderState, bool) {
	fail := func() (tokenizer.ReaderState, bool) { return 0, false }

	switch tokenType {
	case ImageSpanInline:
		return r.Token(s, "![")
	case SpanInline:
		return r.Token(s, "[")
	case SpanInline ^ tokenizer.Open, ImageSpanInline ^ tokenizer.Open:
		return r.Token(s, "]")
	case LinkUrlInline:
		return r.Token(s, "(")
	case LinkUrlInline ^ tokenizer.Open:
		return r.Token(s, ")")
	case LinkReferenceInline:
		return r.Token(s, "[")
	case LinkReferenceInline ^ tokenizer.Open:
		return r.Token(s, "]")
	case AutolinkInline:
		return r.Token(s, "<")
	case AutolinkInline ^ tokenizer.Open:
		return r.Token(s, ">")
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
	case EmphasisInline:
		if next, ok := r.Token(s, "{_"); ok {
			return next, ok
		}
		if next, ok := r.Token(s, "_"); ok && !r.HasMask(next, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case EmphasisInline ^ tokenizer.Open:
		if next, ok := r.Token(s, "_}"); ok {
			return next, ok
		}
		if next, ok := r.Token(s, "_"); ok && !r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case StrongInline:
		if next, ok := r.Token(s, "{*"); ok {
			return next, ok
		}
		if next, ok := r.Token(s, "*"); ok && !r.HasMask(next, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case StrongInline ^ tokenizer.Open:
		if next, ok := r.Token(s, "*}"); ok {
			return next, ok
		}
		if next, ok := r.Token(s, "*"); ok && !r.HasMask(s-1, tokenizer.SpaceNewLineByteMask) {
			return next, ok
		}
		return fail()
	case HighlightedInline:
		return r.Token(s, "{=")
	case HighlightedInline ^ tokenizer.Open:
		return r.Token(s, "=}")
	case SuperscriptInline:
		if next, ok := r.Token(s, "{^"); ok {
			return next, ok
		}
		return r.Token(s, "^")
	case SuperscriptInline ^ tokenizer.Open:
		if next, ok := r.Token(s, "^}"); ok {
			return next, ok
		}
		return r.Token(s, "^")
	case SubscriptInline:
		if next, ok := r.Token(s, "{~"); ok {
			return next, ok
		}
		return r.Token(s, "~")
	case SubscriptInline ^ tokenizer.Open:
		if next, ok := r.Token(s, "~}"); ok {
			return next, ok
		}
		return r.Token(s, "~")
	case InsertInline:
		return r.Token(s, "{+")
	case InsertInline ^ tokenizer.Open:
		return r.Token(s, "+}")
	case DeleteInline:
		return r.Token(s, "{-")
	case DeleteInline ^ tokenizer.Open:
		return r.Token(s, "-}")
	case FootnoteReferenceInline:
		return r.Token(s, "[^")
	case FootnoteReferenceInline ^ tokenizer.Open:
		return r.Token(s, "]")
	case EscapedSymbolInline:
		next, ok := r.Token(s, "\\")
		if !ok {
			return fail()
		}
		if r.IsEmpty(next) {
			return fail()
		}
		if next, ok = r.Mask(next, AsciiPunctuationSymbolByteMask); ok {
			return next, ok
		}
		next, ok = r.MaskRepeat(next, tokenizer.SpaceByteMask, 0)
		tokenizer.Assertf(ok, "MaskRepeat must match because minCount is zero")
		return r.Token(next, "\n")
	case RawFormatInline:
		return r.Token(s, "{=")
	case RawFormatInline ^ tokenizer.Open:
		return r.Token(s, "}")
	case SymbolsInline:
		next, ok := r.Token(s, ":")
		if !ok {
			return fail()
		}
		if next, ok = r.MaskRepeat(next, AlphaNumericSymbolByteMask, 0); ok && r.HasToken(next, ":") {
			return next, true
		}
		return fail()
	case SymbolsInline ^ tokenizer.Open:
		return r.Token(s, ":")
	case SmartSymbolInline:
		if next, ok := r.Token(s, "{"); ok {
			return r.Mask(next, SmartSymbolByteMask)
		}
		if next, ok := r.Mask(s, SmartSymbolByteMask); ok {
			if r.HasToken(next, "}") {
				return next + 1, true
			}
			return next, true
		}
		if next, ok := r.Token(s, "..."); ok {
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
