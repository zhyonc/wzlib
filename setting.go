package wzlib

type Setting struct {
	IsLazyLoad       bool
	LocaleRegion     Region
	MSRegion         Region
	MSVersion        int16
	DisableAESCipher bool
	IVKey            [4]byte
	AESKey           [32]byte
}
