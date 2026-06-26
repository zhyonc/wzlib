package wzlib

import (
	"crypto/aes"
	"crypto/cipher"
	"log/slog"
)

//nolint:gochecknoglobals // default fixed constant
var (
	BMSIVKey = [4]byte{0x00, 0x00, 0x00, 0x00} // BMS
	GMSIVKey = [4]byte{0x4D, 0x23, 0xC7, 0x2B} // GMS/KMS/KMST/JMS
	SEAIVKey = [4]byte{0xB9, 0x7D, 0x63, 0xE9} // CMS/CMST/TMS/MSEA

	AESKeyDefault = [32]byte{
		0x13, 0x00, 0x00, 0x00,
		0x08, 0x00, 0x00, 0x00,
		0x06, 0x00, 0x00, 0x00,
		0xB4, 0x00, 0x00, 0x00,
		0x1B, 0x00, 0x00, 0x00,
		0x0F, 0x00, 0x00, 0x00,
		0x33, 0x00, 0x00, 0x00,
		0x52, 0x00, 0x00, 0x00,
	}
)

type aesCipher struct {
	block cipher.Block
	iv    []byte
}

func NewAESCipher(region Region, ivKey [4]byte, aesKey [32]byte) IAESCipher {
	if ivKey[0] == 0 {
		switch region {
		case BMS:
			ivKey = BMSIVKey
		case GMS, KMS, KMST, JMS:
			ivKey = GMSIVKey
		case CMS, CMST, TMS, MSEA:
			ivKey = SEAIVKey
		default:
			slog.Warn("Unknown region", "region", region)
			return nil
		}
	}
	if aesKey[0] == 0 {
		aesKey = AESKeyDefault
	}
	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		slog.Warn("Unknown aes key", "err", err)
		return nil
	}
	iv := make([]byte, aes.BlockSize)
	for i := range 4 {
		copy(iv[4*i:], ivKey[:])
	}
	c := &aesCipher{
		block: block,
		iv:    iv,
	}
	return c
}

// Decrypt implements [IAESCipher].
func (c *aesCipher) Decrypt(buf []byte) {
	c.OFBUpdate(buf)
}

// Encrypt implements [IAESCipher].
func (c *aesCipher) Encrypt(buf []byte) {
	c.OFBUpdate(buf)
}

// OFBUpdate is char __cdecl CAESCipher::OFB_DecUpdate(CAESCipher::AES_ALG_INFO *AlgInfo,char *CipherTxt,unsigned int CipherTxtLen,char *PlainTxt,unsigned int *PlainTxtLen).
// OFBUpdate is char __cdecl CAESCipher::OFB_EncUpdate(CAESCipher::AES_ALG_INFO *AlgInfo,char *PlainTxt,unsigned int PlainTxtLen,char *CipherTxt,unsigned int *CipherTxtLen).
func (c *aesCipher) OFBUpdate(buf []byte) {
	stream := cipher.NewOFB(c.block, c.iv) //nolint:staticcheck // required for legacy OFB compatibility
	stream.XORKeyStream(buf, buf)
}
