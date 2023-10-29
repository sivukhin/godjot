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

func Intersection(masks ...ByteMask) ByteMask {
	result := masks[0]
	for _, mask := range masks[1:] {
		result = result.And(mask)
	}
	return result
}

type TextReader []byte
type ReaderState int

const Unmatched ReaderState = -1

var (
	SpaceByteMask        = NewByteMask([]byte(" \t"))
	SpaceNewLineByteMask = NewByteMask([]byte(" \t\r\n"))
)

func (s ReaderState) Matched() bool {
	return s != Unmatched
}

func (r TextReader) Select(start, end ReaderState) string {
	return string(r[start:end])
}

func (r TextReader) EmptyOrWhiteSpace(s ReaderState) ReaderState {
	next := r.MaskRepeat(s, SpaceNewLineByteMask, 0)
	if !r.Empty(next) {
		return Unmatched
	}
	return next
}
func (r TextReader) Mask(s ReaderState, mask ByteMask) ReaderState {
	if r.HasMask(s, mask) {
		return s + 1
	}
	return Unmatched
}
func (r TextReader) Token(s ReaderState, token string) ReaderState {
	if r.HasToken(s, token) {
		return s + ReaderState(len([]byte(token)))
	}
	return Unmatched
}
func (r TextReader) ByteRepeat(s ReaderState, b byte, count int) ReaderState {
	for !r.Empty(s) {
		if r.HasByte(s, b) {
			s++
			count--
		} else {
			break
		}
	}
	if count <= 0 {
		return s
	}
	return Unmatched
}
func (r TextReader) MaskRepeat(s ReaderState, mask ByteMask, count int) ReaderState {
	for !r.Empty(s) {
		if r.HasMask(s, mask) {
			s++
			count--
		} else {
			break
		}
	}
	if count <= 0 {
		return s
	}
	return Unmatched
}

func (r TextReader) Empty(current ReaderState) bool {
	return int(current) >= len(r) || current < 0
}
func (r TextReader) HasToken(s ReaderState, token string) bool {
	return bytes.HasPrefix(r[s:], []byte(token))
}
func (r TextReader) HasByte(s ReaderState, b byte) bool {
	if r.Empty(s) {
		return false
	}
	return r[s] == b
}
func (r TextReader) HasMask(s ReaderState, mask ByteMask) bool {
	if r.Empty(s) {
		return false
	}
	return mask.Has(r[s])
}

func (r TextReader) Peek(current ReaderState) (byte, bool) {
	if int(current) < len(r) {
		return r[current], true
	}
	return 0, false
}
