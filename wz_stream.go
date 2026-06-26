package wzlib

import (
	"encoding/binary"
	"log/slog"
	"math"
	"os"

	"github.com/edsrzf/mmap-go"
)

type wzStream struct {
	aesCipher IAESCipher
	locale    ILocale
	data      mmap.MMap
	length    int64
	offset    int64
}

func NewWzStream(path string, aesCipher IAESCipher, locale ILocale) (IWzStream, error) {
	s := &wzStream{
		aesCipher: aesCipher,
		locale:    locale,
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()
	data, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	s.data = data
	s.length = int64(len(s.data))
	return s, nil
}

// GetRemain implements [IWzStream].
func (s *wzStream) GetRemain() int64 {
	return s.length - s.offset
}

// GetLength implements [IWzStream].
func (s *wzStream) GetLength() int64 {
	return s.length
}

// GetOffset implements [IWzStream].
func (s *wzStream) GetOffset() int64 {
	return s.offset
}

// SetOffset implements [IWzStream].
func (s *wzStream) SetOffset(offset int64) {
	s.offset = offset
}

// DecryptDataOffset implements [IWzStream].
func (s *wzStream) DecryptDataOffset(headerSize int32, version int16) int32 {
	hash := GetOffsetHash(int32(s.offset), headerSize, version)
	hashXOR := uint32(s.Decode4())
	xor := hashXOR ^ hash
	dataOffset := xor + (uint32(headerSize) * 2)
	return int32(dataOffset)
}

// DecodeBool implements [IWzStream].
func (s *wzStream) DecodeBool() bool {
	return s.Decode1() == 1
}

// Decode1 implements [IWzStream].
func (s *wzStream) Decode1() int8 {
	if s.offset+1 > s.length {
		return 0
	}
	val := int8(s.data[s.offset])
	s.offset++
	return val
}

// Decode2 implements [IWzStream].
func (s *wzStream) Decode2() int16 {
	if s.offset+2 > s.length {
		return 0
	}
	val := int16(binary.LittleEndian.Uint16(s.data[s.offset : s.offset+2]))
	s.offset += 2
	return val
}

// Decode4 implements [IWzStream].
func (s *wzStream) Decode4() int32 {
	if s.offset+4 > s.length {
		return 0
	}
	val := int32(binary.LittleEndian.Uint32(s.data[s.offset : s.offset+4]))
	s.offset += 4
	return val
}

// Decode8 implements [IWzStream].
func (s *wzStream) Decode8() int64 {
	if s.offset+8 > s.length {
		return 0
	}
	val := int64(binary.LittleEndian.Uint64(s.data[s.offset : s.offset+8]))
	s.offset += 8
	return val
}

// Decode8f implements [IWzStream].
func (s *wzStream) Decode8f() float64 {
	return math.Float64frombits(uint64(s.Decode8()))
}

// DecodeVTLen implements [IWzStream].
func (s *wzStream) DecodeVTLen() (int8, bool) {
	b := s.Decode1()
	if b == -math.MaxInt8-1 { // -128(0x80)
		return b, true
	}
	return b, false
}

// DecodeVT4 implements [IWzStream].
func (s *wzStream) DecodeVT4() int32 {
	b, isVTLen := s.DecodeVTLen()
	if isVTLen {
		return s.Decode4()
	}
	return int32(b)
}

// DecodeVT4f implements [IWzStream].
func (s *wzStream) DecodeVT4f() float32 {
	b, isVTLen := s.DecodeVTLen()
	if isVTLen {
		return math.Float32frombits(uint32(s.Decode4()))
	}
	return float32(b)
}

// DecodeVT8 implements [IWzStream].
func (s *wzStream) DecodeVT8() int64 {
	b, isVTLen := s.DecodeVTLen()
	if isVTLen {
		return s.Decode8()
	}
	return int64(b)
}

// DecodeBuffer implements [IWzStream].
func (s *wzStream) DecodeBuffer(bufLen int64) []byte {
	remain := s.GetRemain()
	if bufLen <= 0 || remain <= 0 {
		return nil
	}
	if remain < bufLen {
		bufLen = remain
	}
	buf := make([]byte, bufLen)
	copy(buf, s.data[s.offset:s.offset+bufLen])
	s.offset += bufLen
	return buf
}

// DecryptBuffer implements [IWzStream].
func (s *wzStream) DecryptBuffer(bufLen int64) []byte {
	buf := s.DecodeBuffer(bufLen)
	if len(buf) == 0 {
		return buf
	}
	if s.aesCipher != nil {
		s.aesCipher.Decrypt(buf)
	}
	return buf
}

// DecodeStr implements [IWzStream].
func (s *wzStream) DecodeStr(strLen int64) string {
	if strLen <= 0 {
		return ""
	}
	if s.offset+strLen > s.length {
		strLen = s.GetRemain()
	}
	data := s.data[s.offset : s.offset+strLen]
	s.offset += strLen
	return string(data)
}

// DecodeNTStr implements [IWzStream].
func (s *wzStream) DecodeNTStr() string {
	var buf []byte
	for s.offset < s.length {
		b := s.data[s.offset]
		s.offset++
		if b == 0x00 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf)
}

// DecodeVTStrLen implements [IWzStream].
func (s *wzStream) DecodeVTStrLen() (int32, bool) {
	strLen := int32(s.Decode1())
	switch {
	case strLen >= -math.MaxInt8 && strLen < 0:
		// ASCII Str range [-127,0)
		return -strLen, true
	case strLen == -math.MaxInt8-1:
		// Variant ASCII Str set -128(0x80)
		return s.Decode4(), true
	case strLen > 0 && strLen < math.MaxInt8:
		// UTF‑16LE Str range (0,127)
		return strLen * 2, false
	case strLen == math.MaxInt8:
		// Variant UTF‑16LE Str set 127
		return s.Decode4() * 2, false
	default:
		return 0, false
	}
}

// DecodeVTStr implements [IWzStream].
func (s *wzStream) DecodeVTStr() string {
	strLen, isASCII := s.DecodeVTStrLen()
	if strLen == 0 {
		return ""
	}
	buf := s.DecodeBuffer(int64(strLen))
	if isASCII {
		return s.locale.Decode(buf)
	}
	return s.locale.DecodeUTF16(buf)
}

// DecryptVTStr implements [IWzStream].
func (s *wzStream) DecryptVTStr() string {
	strLen, isASCII := s.DecodeVTStrLen()
	if strLen == 0 {
		return ""
	}
	// XOR Cipher
	buf := s.DecodeBuffer(int64(strLen))
	if isASCII {
		var mask uint8 = 0xAA
		for i, b := range buf {
			buf[i] = b ^ mask
			mask++
		}
		if s.aesCipher == nil {
			return s.locale.Decode(buf)
		}
	} else {
		var mask uint16 = 0xAAAA
		for i := 0; i < len(buf); i += 2 {
			char := binary.LittleEndian.Uint16(buf[i:])
			char ^= mask
			binary.LittleEndian.PutUint16(buf[i:], char)
			mask++
		}
		if s.aesCipher == nil {
			return s.locale.DecodeUTF16(buf)
		}
	}
	// AES Cipher
	s.aesCipher.Decrypt(buf)
	if isASCII {
		return s.locale.Decode(buf)
	}
	return s.locale.DecodeUTF16(buf)
}

// DecryptVTStrRef implements [IWzStream].
func (s *wzStream) DecryptVTStrRef(ref int32, dataOffset int32) string {
	var str string
	backOffset := s.offset
	s.offset = int64(dataOffset + ref)
	str = s.DecryptVTStr()
	s.offset = backOffset
	return str
}

// Back implements [IWzStream].
func (s *wzStream) Back(n int64) {
	if n <= 0 {
		return
	}
	if n <= s.offset {
		s.offset -= n
	} else {
		s.offset = 0
	}
}

// Skip implements [IWzStream].
func (s *wzStream) Skip(n int64) {
	if n <= 0 {
		return
	}
	remain := s.GetRemain()
	if n < remain {
		s.offset += n
	} else {
		s.offset = s.length
	}
}

// Close implements [IWzStream].
func (s *wzStream) Close() {
	if s.data != nil {
		err := s.data.Unmap()
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}
}
