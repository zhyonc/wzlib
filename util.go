package wzlib

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/bits"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// #region File

func GetFilePaths(dir string, ext string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check info
		if info == nil {
			return fmt.Errorf("file info is nil for path: %s", path)
		}
		if info.IsDir() {
			return nil
		}
		// Check file ext format
		if filepath.Ext(info.Name()) != ext {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func CreateDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0750)
}

func WriteFile(filePath string, buf []byte) error {
	return os.WriteFile(filePath, buf, 0600)
}

func SaveFile(filePath string, buf []byte) error {
	err := CreateDir(filePath)
	if err != nil {
		return err
	}
	return WriteFile(filePath, buf)
}

// #endregion

// #region HashSum

func GetOffsetHash(offset int32, headerSize int32, version int16) uint32 {
	hash := uint32(offset) - uint32(headerSize)
	hash = ^hash
	hash *= GetWzVersionHash(version)
	hash -= 0x581C3F6D  // magic number
	mask := hash & 0x1F // bitmask slicing
	hash = bits.RotateLeft32(hash, int(mask))
	return hash
}

func GetWzVersionHash(version int16) uint32 {
	var hash uint32
	// Cast the base 10 version into an ascii string
	str := strconv.Itoa(int(version))
	for _, b := range str {
		// hash = 32*hash + str[i] + 1
		hash = (hash << 5) + uint32(b) + 1
	}
	return hash
}

func GetWzVersionHashSum(version int16) int16 {
	hash := GetWzVersionHash(version)
	sum := 0xFF ^ (hash >> 24) ^ (hash >> 16) ^ (hash >> 8) ^ hash
	return int16(sum)
}

func GetWzVersionHashSumTruncated(version int16) int16 {
	return GetWzVersionHashSum(version) & 0xFF
}

func GetPossibleVersions(versionHashSumTruncated int16) []int16 {
	if versionHashSumTruncated == 0 {
		return nil
	}
	possibleVersions := make([]int16, 0)
	for version := range math.MaxUint16 {
		hashSumTruncated := GetWzVersionHashSumTruncated(int16(version))
		if uint16(hashSumTruncated) == uint16(versionHashSumTruncated) {
			if version >= MinVersion && version <= MaxVersion {
				possibleVersions = append(possibleVersions, int16(version))
			}
		}
	}
	return possibleVersions
}

func CalcuCheckSum(buf []byte) int32 {
	var sum int32
	for _, b := range buf {
		sum += int32(b)
	}
	return sum
}

// #endregion

// #region String

func IsJSON(s string) bool {
	strLen := len(s)
	if strLen >= 2 {
		first := s[0]
		last := s[strLen-1]
		return first == '{' && last == '}'
	}
	return false
}

func IsJSONArray(s string) bool {
	strLen := len(s)
	switch {
	case strLen == 2:
		return s[0] == '[' && s[1] == ']'
	case strLen >= 4:
		first := s[0]
		last := s[strLen-1]
		second := s[1]
		secondLast := s[strLen-2]
		return first == '[' && last == ']' && second == '{' && secondLast == '}'
	}
	return false
}

func IsASCII(s string) bool {
	if IsJSON(s) || IsJSONArray(s) {
		return true
	}
	for _, r := range s {
		if r > math.MaxInt8 {
			_, ok := ASCIIWhitelist[r]
			if !ok {
				return false
			}
		}
	}
	return true
}

func DumpString(buf []byte) string {
	bufLen := len(buf)
	var builder strings.Builder
	for i, b := range buf {
		fmt.Fprintf(&builder, "%02X", b)
		if i < bufLen-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

func GetFontExt(data []byte) string {
	if len(data) < 4 {
		return ".bin"
	}
	if binary.BigEndian.Uint32(data[:4]) == 0x10000 {
		return ".ttf"
	}
	if string(data[:4]) == "OTTO" {
		return ".otf"
	}
	return ".bin"
}

// #endregion

// #region Zlib

func IsZlibHeader(header uint16) bool {
	zHeader := ZlibHeader(header)
	switch zHeader {
	case ZlibNoCompressionHeader,
		ZlibBestSpeedHeader,
		ZlibBestCompressionHeader,
		ZlibDefaultCompressionHeader:
		return true
	default:
		return false
	}
}

func ZlibInflate(data []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	raw, err := io.ReadAll(zr)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			// Miss Adler32 CheckSum
			return raw, nil
		}
		return nil, err
	}

	return raw, nil
}

func ZlibDeflate(raw []byte, level ZlibLevel) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, int(level))
	if err != nil {
		return nil, err
	}
	_, err = zw.Write(raw)
	if err != nil {
		return nil, err
	}

	err = zw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// #endregion
