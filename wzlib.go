package wzlib

import (
	"context"
	"image"
	"time"
)

type IAESCipher interface {
	Decrypt(buf []byte)
	Encrypt(buf []byte)
}

type ILocale interface {
	Decode(buf []byte) string
	Encode(s string) []byte
	DecodeUTF16(buf []byte) string
	EncodeUTF16(s string) []byte
}

type IWzFile interface {
	IWzRootNode
	GetNode(nodePath string) (IWzNode, error)
	GetDirectory(nodePath string) (IWzDirectory, error)
	GetImage(nodePath string) (IWzImage, error)
	GetPropertyItem(nodePath string) (IWzPropertyItem, error)
	GetMetaNodeLen() int
	GetMetaNodeKeys() []string
	Elapsed() time.Duration
	Load(filePath string) error
	ReadHeader(stream IWzStream) error
	PredictVersion(stream IWzStream) error
	ReadMetaNode(ctx context.Context, stream IWzStream, parent IWzNode) error
	ReadDataNode(ctx context.Context, stream IWzStream) error
	Save(filePath string, isCreate bool) error
	WriteHeader(archive IWzArchive) error
	WriteMetaNode(ctx context.Context, archive IWzArchive, parent IWzNode) error
	WriteDataNode(ctx context.Context, archive IWzArchive) error
	Close()
}

type IWzStream interface {
	GetRemain() int64
	GetLength() int64
	GetOffset() int64
	SetOffset(offset int64)
	DecryptDataOffset(headerSize int32, version int16) int32
	DecodeBool() bool
	Decode1() int8
	Decode2() int16
	Decode4() int32
	Decode8() int64
	Decode8f() float64
	DecodeVTLen() (int8, bool)
	DecodeVT4() int32
	DecodeVT4f() float32
	DecodeVT8() int64
	DecodeBuffer(bufLen int64) []byte
	DecryptBuffer(bufLen int64) []byte
	DecryptBlock(block []byte) []byte
	DecodeStr(strLen int64) string
	DecodeNTStr() string
	DecodeVTStrLen() (int32, bool)
	DecodeVTStr() string
	DecryptVTStr() string
	DecryptVTStrRef(ref int32, dataOffset int32) string
	Back(n int64)
	Skip(n int64)
	Close()
}

type IWzArchive interface {
	GetBuffer() []byte
	GetLength() int64
	GetOffset() int64
	SetOffset(offset int64)
	EncryptDataOffset(dataOffset int32, headerSize int32, version int16)
	EncodeBool(b bool)
	Encode1(n int8)
	Encode2(n int16)
	Encode4(n int32)
	Encode8(n int64)
	Encode8f(n float64)
	EncodeVTLen(n any) bool
	EncodeVT4(n int32)
	EncodeVT4f(n float32)
	EncodeVT8(n int64)
	EncodeBuffer(buf []byte)
	EncryptBuffer(buf []byte)
	EncodeStr(s string)
	EncodeNTStr(s string)
	EncodeVTStrLen(s string) bool
	EncodeVTStr(s string)
	EncryptVTStr(s string)
	EncryptVTStrRef(ref int32)
	Back(n int64)
	Close()
}

type IWzSerialize interface {
	DeSerialize(stream IWzStream)
	Serialize(archive IWzArchive)
	GetSelfNode() IWzNode
}

type IWzRootNode interface {
	GetIdent() string
	GetSize() int64
	GetHeaderSize() int32
	GetCopyright() string
	GetVersion() int16
	GetStream() IWzStream
}

//nolint:iface // interfaces kept separate for semantic clarity
type IWzNode interface {
	IWzSerialize
	GetType() NodeType
	SetType(nType NodeType)
	GetFlag() FlagType
	SetFlag(flag FlagType)
	GetName() string
	SetName(name string)
	GetNameRef() int32
	SetNameRef(offset int32)
	GetSize() int32
	SetSize(size int32)
	GetCheckSum() int32
	SetCheckSum(sum int32)
	GetOffset() int32
	SetOffset(offset int32)
	GetRootNode() IWzRootNode
	SetRootNode(root IWzRootNode)
	GetMetaNode() IWzNode
	GetDataOffset() int32
	Traversal(paths []string, stream IWzStream) (IWzNode, error)
	TraverslChild(paths []string) (IWzNode, error)
	GetParent() IWzNode
	GetChildsLen() int32
	GetChildNames() []string
	GetChilds() []IWzNode
	GetFirstChild() IWzNode
	GetChild(name string) (IWzNode, error)
	GetChildByPath(nodePath string) (IWzNode, error)
	AddChild(node IWzNode)
	ParseDirectory() (IWzDirectory, error)
	ParseImage() (IWzImage, error)
	ParseItem() (IWzPropertyItem, error)
}

type IWzDirectory interface {
	IWzNode
	GetDirectory(nodePath string) (IWzDirectory, error)
	GetImage(nodePath string) (IWzImage, error)
}

type IWzImage interface {
	IWzNode
	GetPropertyItem(nodePath string) (IWzPropertyItem, error)
}

//nolint:iface // interfaces kept separate for semantic clarity
type IWzProperty interface {
	IWzNode
}

type IWzPropertyItem interface {
	IWzNode
	GetValueType() VARENUM
	SetValueType(vt VARENUM)
	GetValue() any
	SetValue(value any)
	SetVTStr(str string)
	SetVTStrRef(strRef int32)
	GetInt16() (int16, error)
	GetInt32() (int32, error)
	GetInt64() (int64, error)
	GetFloat32() (float32, error)
	GetFloat64() (float64, error)
	GetString() (string, error)
	GetCanvas() (IWzCanvas, error)
	GetVector() (IWzVector, error)
	GetConvex() (IWzConvex, error)
	GetSound() (IWzSound, error)
	GetUOL() (IWzUOL, error)
	GetRawData() (IWzRawData, error)
	GetScript() (IWzLua, error)
}

type IWzCanvas interface {
	IWzNode
	GetProperty() IWzProperty
	SetProperty(prop IWzProperty)
	GetWidth() int32
	SetWidth(width int32)
	GetHeight() int32
	SetHeight(height int32)
	GetPixelFormat() CanvasPixformat
	SetPixelFormat(cp CanvasPixformat)
	GetMagLevel() int8
	SetMagLevel(magLevel int8)
	GetData() []byte
	SetData(data []byte)
	GetRawData() ([]byte, error)
	ExtractImage() (*image.RGBA, error)
}

type IWzVector interface {
	IWzNode
	GetX() int32
	SetX(x int32)
	GetY() int32
	SetY(y int32)
}

type IWzConvex interface {
	IWzNode
	GetFoothold(index int) IWzVector
	SetFoothold(index int, foothold IWzVector)
	GetFootholds() []IWzVector
	SetFootholds(footholds []IWzVector)
}

type IWzUOL interface {
	IWzNode
	GetUOL() string
	SetUOL(uol string)
}

type IWzSound interface {
	IWzNode
	GetDuration() int32
	SetDuration(duration int32)
	GetSoundType() SoundType
	SetSoundType(soundType SoundType)
	GetMediaType() []byte
	SetMediaType(mediaType []byte)
	GetMediaSubType() []byte
	SetMediaSubType(mediaSubType []byte)
	GetFormatType() []byte
	SetFormatType(formatType []byte)
	GetFormat() any
	SetFormat(format any)
	GetData() []byte
	SetData(data []byte)
	GetRawData() ([]byte, error)
	ExtractAudio() ([]byte, error) // MP3/WAV data
	ExtractBlob() ([]byte, error)  // Font/Spine data
}

//nolint:iface // interfaces kept separate for semantic clarity
type IWzRawData interface {
	IWzNode
	GetData() []byte
	SetData(data []byte)
}

//nolint:iface // interfaces kept separate for semantic clarity
type IWzLua interface {
	IWzNode
	GetData() []byte
	SetData(data []byte)
}
