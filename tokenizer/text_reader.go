package tokenizer

import (
	"bytes"
)

type ByteMask [4]uint64

func NewByteMask(set []byte) ByteMask {
	mask := ByteMask{0, 0}
	for _, b := range set {
		mask[b/64] |= 1 << (b & 63)
	}
	return mask
}
func (m ByteMask) Has(b byte) bool {
	return m[b/64]&(1<<uint64(b&63)) > 0
}

func (m ByteMask) Negate() ByteMask {
	const mask uint64 = 0xffffffffffffffff
	return ByteMask{m[0] ^ mask, m[1] ^ mask, m[2] ^ mask, m[3] ^ mask}
}

func (m ByteMask) Or(other ByteMask) ByteMask {
	return ByteMask{m[0] | other[0], m[1] | other[1], m[2] | other[2], m[3] | other[3]}
}

func (m ByteMask) And(other ByteMask) ByteMask {
	return ByteMask{m[0] & other[0], m[1] & other[1], m[2] & other[2], m[3] & other[3]}
}

func Union(masks ...ByteMask) ByteMask {
	var result ByteMask
	for _, mask := range masks {
		result = result.Or(mask)
	}
	return result
}

type (
	TextReader  []byte
	ReaderState = int
)

var (
	SpaceByteMask        = NewByteMask([]byte(" \t"))
	SpaceNewLineByteMask = NewByteMask([]byte(" \t\r\n"))
)

func (r TextReader) Select(start, end ReaderState) string {
	return string(r[start:end])
}
func (r TextReader) EmptyOrWhiteSpace(s ReaderState) (ReaderState, bool) {
	next, _ := r.MaskRepeat(s, SpaceNewLineByteMask, 0)
	if !r.IsEmpty(next) {
		return 0, false
	}
	return next, true
}
func (r TextReader) Mask(s ReaderState, mask ByteMask) (ReaderState, bool) {
	if r.HasMask(s, mask) {
		return s + 1, true
	}
	return 0, false
}
func (r TextReader) Token(s ReaderState, token string) (ReaderState, bool) {
	if r.HasToken(s, token) {
		return s + len([]byte(token)), true
	}
	return 0, false
}
func (r TextReader) ByteRepeat(s ReaderState, b byte, minCount int) (ReaderState, bool) {
	for !r.IsEmpty(s) {
		if r.HasByte(s, b) {
			s++
			minCount--
		} else {
			break
		}
	}
	if minCount <= 0 {
		return s, true
	}
	return 0, false
}
func (r TextReader) MaskRepeat(s ReaderState, mask ByteMask, minCount int) (ReaderState, bool) {
	for !r.IsEmpty(s) {
		if r.HasMask(s, mask) {
			s++
			minCount--
		} else {
			break
		}
	}
	if minCount <= 0 {
		return s, true
	}
	return 0, false
}

func (r TextReader) IsEmptyOrWhiteSpace(s ReaderState) bool {
	next, _ := r.MaskRepeat(s, SpaceNewLineByteMask, 0)
	return r.IsEmpty(next)
}
func (r TextReader) IsEmpty(s ReaderState) bool {
	return s >= len(r) || s < 0
}
func (r TextReader) HasToken(s ReaderState, token string) bool {
	return bytes.HasPrefix(r[s:], []byte(token))
}
func (r TextReader) HasByte(s ReaderState, b byte) bool {
	if r.IsEmpty(s) {
		return false
	}
	return r[s] == b
}
func (r TextReader) HasMask(s ReaderState, mask ByteMask) bool {
	if r.IsEmpty(s) {
		return false
	}
	return mask.Has(r[s])
}
func (r TextReader) Peek(s ReaderState) (byte, bool) {
	if s < len(r) {
		return r[s], true
	}
	return 0, false
}
