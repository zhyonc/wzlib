package wzlib

import (
	"encoding/binary"
	"math"
	"unicode/utf8"
)

type wzArchive struct {
	aesCipher IAESCipher
	locale    ILocale
	data      []byte
	offset    int64
}

func NewWzArchive(aesCipher IAESCipher, locale ILocale) IWzArchive {
	a := &wzArchive{
		aesCipher: aesCipher,
		locale:    locale,
	}
	a.data = make([]byte, 0)
	return a
}

// GetBuffer implements [IWzArchive].
func (a *wzArchive) GetBuffer() []byte {
	return a.data
}

// GetLength implements [IWzArchive].
func (a *wzArchive) GetLength() int64 {
	return int64(len(a.data))
}

// GetOffset implements [IWzArchive].
func (a *wzArchive) GetOffset() int64 {
	return a.offset
}

// SetOffset implements [IWzArchive].
func (a *wzArchive) SetOffset(offset int64) {
	a.offset = offset
}

// EncryptDataOffset implements [IWzArchive].
func (a *wzArchive) EncryptDataOffset(dataOffset int32, headerSize int32, version int16) {
	hash := GetOffsetHash(int32(a.offset), headerSize, version)
	xor := uint32(dataOffset) - (uint32(headerSize) * 2)
	hashXOR := hash ^ xor
	a.Encode4(int32(hashXOR))
}

// EncodeBool implements [IWzArchive].
func (a *wzArchive) EncodeBool(b bool) {
	var n int8
	if b {
		n = 1
	}
	a.Encode1(n)
}

// Encode1 implements [IWzArchive].
func (a *wzArchive) Encode1(n int8) {
	a.data = append(a.data, byte(n))
	a.offset++
}

// Encode2 implements [IWzArchive].
func (a *wzArchive) Encode2(n int16) {
	a.data = append(a.data, byte(n), byte(n>>8))
	a.offset += 2
}

// Encode4 implements [IWzArchive].
func (a *wzArchive) Encode4(n int32) {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(n))
	a.data = append(a.data, tmp...)
	a.offset += 4
}

// Encode8 implements [IWzArchive].
func (a *wzArchive) Encode8(n int64) {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, uint64(n))
	a.data = append(a.data, tmp...)
	a.offset += 8
}

// Encode8f implements [IWzArchive].
func (a *wzArchive) Encode8f(n float64) {
	a.Encode8(int64(math.Float64bits(n)))
}

// EncodeVTLen implements [IWzArchive].
func (a *wzArchive) EncodeVTLen(n any) bool {
	switch v := n.(type) {
	case int32:
		if v > -math.MaxInt8-1 && v <= math.MaxInt8 {
			// (-128,127]
			a.Encode1(int8(v))
			return false
		}
		a.Encode1(-math.MaxInt8 - 1) // -128(0x80)
		return true
	case float32:
		if v == float32(0) {
			a.Encode1(0)
			return false
		}
		a.Encode1(-math.MaxInt8 - 1) // -128(0x80)
		return true
	case int64:
		if v > -math.MaxInt8-1 && v <= math.MaxInt8 {
			// (-128,127]
			a.Encode1(int8(v))
			return false
		}
		a.Encode1(-math.MaxInt8 - 1) // -128(0x80)
		return true
	default:
		a.Encode1(0)
		return false
	}
}

// EncodeVT4 implements [IWzArchive].
func (a *wzArchive) EncodeVT4(n int32) {
	isVTLen := a.EncodeVTLen(n)
	if isVTLen {
		a.Encode4(n)
	}
}

// EncodeVT4f implements [IWzArchive].
func (a *wzArchive) EncodeVT4f(n float32) {
	isVTLen := a.EncodeVTLen(n)
	if isVTLen {
		a.Encode4(int32(math.Float32bits(n)))
	}
}

// EncodeVT8 implements [IWzArchive].
func (a *wzArchive) EncodeVT8(n int64) {
	isVTLen := a.EncodeVTLen(n)
	if isVTLen {
		a.Encode8(n)
	}
}

// EncodeBuffer implements [IWzArchive].
func (a *wzArchive) EncodeBuffer(buf []byte) {
	if len(buf) == 0 {
		return
	}
	a.data = append(a.data, buf...)
	a.offset += int64(len(buf))
}

// EncryptBuffer implements [IWzArchive].
func (a *wzArchive) EncryptBuffer(buf []byte) {
	if a.aesCipher != nil {
		a.aesCipher.Encrypt(buf)
	}
	a.EncodeBuffer(buf)
}

// EncodeStr implements [IWzArchive].
func (a *wzArchive) EncodeStr(s string) {
	tmp := []byte(s)
	a.EncodeBuffer(tmp)
}

// EncodeNTStr implements [IWzArchive].
func (a *wzArchive) EncodeNTStr(s string) {
	a.EncodeStr(s)
	a.Encode1(0)
}

// EncodeVTStrLen implements [IWzArchive].
func (a *wzArchive) EncodeVTStrLen(s string) bool {
	strLen := utf8.RuneCountInString(s) // actual rune length
	if IsASCII(s) {
		if strLen <= math.MaxInt8 {
			// ASCII Str range (0,127]
			a.Encode1(int8(-strLen))
		} else {
			// Variant ASCII Str set -128(0x80)
			a.Encode1(-math.MaxInt8 - 1)
			a.Encode4(int32(strLen))
		}
		return true
	}
	// UTF-16LE
	if strLen < math.MaxInt8 {
		// UTF‑16LE Str range (0,127)
		a.Encode1(int8(strLen))
	} else {
		// Variant UTF‑16LE Str set 127
		a.Encode1(math.MaxInt8)
		a.Encode4(int32(strLen))
	}
	return false
}

// EncodeVTStr implements [IWzArchive].
func (a *wzArchive) EncodeVTStr(s string) {
	if len(s) == 0 {
		return
	}
	isASCII := a.EncodeVTStrLen(s)
	if isASCII {
		a.EncodeBuffer(a.locale.Encode(s))
	} else {
		a.EncodeBuffer(a.locale.EncodeUTF16(s))
	}
}

// EncryptVTStr implements [IWzArchive].
func (a *wzArchive) EncryptVTStr(s string) {
	if len(s) == 0 {
		a.Encode1(0)
		return
	}
	isASCII := a.EncodeVTStrLen(s)
	var buf []byte
	if isASCII {
		buf = a.locale.Encode(s)
		// AES Cipher
		if a.aesCipher != nil {
			a.aesCipher.Encrypt(buf)
		}
		// XOR Cipher
		var mask uint8 = 0xAA
		for i, b := range buf {
			buf[i] = b ^ mask
			mask++
		}
	} else {
		buf = a.locale.EncodeUTF16(s)
		// AES Cipher
		if a.aesCipher != nil {
			a.aesCipher.Encrypt(buf)
		}
		// XOR Cipher
		var mask uint16 = 0xAAAA
		for i := 0; i < len(buf); i += 2 {
			char := binary.LittleEndian.Uint16(buf[i:])
			char ^= mask
			binary.LittleEndian.PutUint16(buf[i:], char)
			mask++
		}
	}
	a.EncodeBuffer(buf)
}

// EncryptVTStrRef implements [IWzArchive].
func (a *wzArchive) EncryptVTStrRef(ref int32) {
	a.Encode4(ref)
}

// Back implements [IWzArchive].
func (a *wzArchive) Back(n int64) {
	if n <= 0 {
		return
	}
	if n <= a.offset {
		a.offset -= n
	} else {
		a.offset = 0
	}
}

// Close implements [IWzArchive].
func (a *wzArchive) Close() {
	a.data = nil
	a.offset = 0
}
