package wzlib

import (
	"bytes"
	"log/slog"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type locale struct {
	decoder        *encoding.Decoder
	encoder        *encoding.Encoder
	utf16LEDecoder *encoding.Decoder
	utf16LEEncoder *encoding.Encoder
}

func NewLocale(region Region) ILocale {
	utf16LE := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	l := &locale{
		utf16LEDecoder: utf16LE.NewDecoder(),
		utf16LEEncoder: utf16LE.NewEncoder(),
	}
	switch region {
	case KMS, KMST:
		l.decoder = korean.EUCKR.NewDecoder()
		l.encoder = korean.EUCKR.NewEncoder()
	case JMS:
		l.decoder = japanese.ShiftJIS.NewDecoder()
		l.encoder = japanese.ShiftJIS.NewEncoder()
	case CMS, CMST, MSEA:
		l.decoder = simplifiedchinese.GBK.NewDecoder()
		l.encoder = simplifiedchinese.GBK.NewEncoder()
	case TMS:
		l.decoder = traditionalchinese.Big5.NewDecoder()
		l.encoder = traditionalchinese.Big5.NewEncoder()
	case GMS, BMS:
		l.decoder = charmap.Windows1252.NewDecoder()
		l.encoder = charmap.Windows1252.NewEncoder()
	default:
		l.decoder = unicode.UTF8.NewDecoder()
		l.encoder = unicode.UTF8.NewEncoder()
	}
	return l
}

// Decode implements [ILocale].
func (lc *locale) Decode(buf []byte) string {
	if len(buf) == 0 {
		return ""
	}
	if before, _, ok := bytes.Cut(buf, []byte{0}); ok {
		buf = before
	}
	result, _, err := transform.Bytes(lc.decoder, buf)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	return string(result)
}

// Encode implements [ILocale].
func (lc *locale) Encode(s string) []byte {
	if s == "" {
		return nil
	}
	result, _, err := transform.String(lc.encoder, s)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	return []byte(result)
}

// DecodeUTF16 implements [ILocale].
func (lc *locale) DecodeUTF16(buf []byte) string {
	if len(buf) == 0 {
		return ""
	}
	result, _, err := transform.Bytes(lc.utf16LEDecoder, buf)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	return string(result)
}

// EncodeUTF16 implements [ILocale].
func (lc *locale) EncodeUTF16(s string) []byte {
	if s == "" {
		return nil
	}
	result, _, err := transform.String(lc.utf16LEEncoder, s)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	return []byte(result)
}
