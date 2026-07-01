package wzlib

import "compress/flate"

// #region Region

type Region uint8

const (
	GMSCW Region = 1
	KMS   Region = 1
	KMST  Region = 2
	JMS   Region = 3
	CMS   Region = 4
	CMST  Region = 5
	TMS   Region = 6
	MSEA  Region = 7
	GMS   Region = 8
	BMS   Region = 9
)

// #endregion

// #region NodeType

type NodeType uint8

const (
	NodeImageRef NodeType = iota + 2
	NodeDirectory
	NodeImage
	NodeProperty
	NodePropertyItem
	NodeCanvas
	NodeVector2D
	NodeConvex2D
	NodeSoundDX8
	NodeUOL
	NodeRawData
	NodeLua
)

// #endregion

// #region NodeNameTag

type NodeNameTag string

const (
	NodeNameProperty NodeNameTag = "Property"
	NodeNameCanvas   NodeNameTag = "Canvas"
	NodeNameVector2D NodeNameTag = "Shape2D#Vector2D"
	NodeNameConvex2D NodeNameTag = "Shape2D#Convex2D"
	NodeNameSoundDX8 NodeNameTag = "Sound_DX8"
	NodeNameUOL      NodeNameTag = "UOL"
	NodeNameRawData  NodeNameTag = "RawData"
)

// #endregion

// #region FlagType

type FlagType uint8

const (
	FlagVTStr       FlagType = 0x00
	FlagVTStrRef    FlagType = 0x01
	FlagPropName    FlagType = 0x73
	FlagPropNameRef FlagType = 0x1B
)

// #endregion

// #region VARENUM

type VARENUM int32

const (
	VTEmpty           VARENUM = 0x0 // use
	VTNull            VARENUM = 0x1 // use
	VTI2              VARENUM = 0x2 // use
	VTI4              VARENUM = 0x3 // use
	VTR4              VARENUM = 0x4 // use
	VTR8              VARENUM = 0x5 // use
	VTCy              VARENUM = 0x6
	VTDate            VARENUM = 0x7
	VTBStr            VARENUM = 0x8 // use
	VTDispatch        VARENUM = 0x9 // use
	VTError           VARENUM = 0xA
	VTBool            VARENUM = 0xB // use
	VTVariant         VARENUM = 0xC
	VTUnknown         VARENUM = 0xD
	VTDecimal         VARENUM = 0xE
	VTI1              VARENUM = 0x10
	VTUI1             VARENUM = 0x11
	VTUI2             VARENUM = 0x12 // use
	VTUI4             VARENUM = 0x13 // use
	VTI8              VARENUM = 0x14 // use
	VTUI8             VARENUM = 0x15 // use
	VTInt             VARENUM = 0x16 // use
	VTUInt            VARENUM = 0x17 // use
	VTVoid            VARENUM = 0x18
	VTHResult         VARENUM = 0x19
	VTPtr             VARENUM = 0x1A
	VTSafeArray       VARENUM = 0x1B
	VTCarray          VARENUM = 0x1C
	VTUserDefined     VARENUM = 0x1D
	VTLPStr           VARENUM = 0x1E
	VTLPWStr          VARENUM = 0x1F
	VTRecord          VARENUM = 0x24
	VTIntPtr          VARENUM = 0x25
	VTUIntPtr         VARENUM = 0x26
	VTFileTime        VARENUM = 0x40
	VTBlob            VARENUM = 0x41
	VTStream          VARENUM = 0x42
	VTStorage         VARENUM = 0x43
	VTStreamedObject  VARENUM = 0x44
	VTStoredObject    VARENUM = 0x45
	VTBlobObject      VARENUM = 0x46
	VTCF              VARENUM = 0x47
	VTClsID           VARENUM = 0x48
	VTVersionedStream VARENUM = 0x49
	VTBStrBlob        VARENUM = 0xFFF
	VTVector          VARENUM = 0x1000
	VTArray           VARENUM = 0x2000
	VTByRef           VARENUM = 0x4000
	VTReserved        VARENUM = 0x8000
	VTIllegal         VARENUM = 0xFFFF
	VTIllegalMasked   VARENUM = 0xFFF
	VTTypeMask        VARENUM = 0xFFF
)

// #endregion

// #region Zlib

type ZlibHeader uint16

const (
	ZlibNoCompressionHeader      ZlibHeader = 0x7801
	ZlibBestSpeedHeader          ZlibHeader = 0x785E
	ZlibBestCompressionHeader    ZlibHeader = 0x78DA
	ZlibDefaultCompressionHeader ZlibHeader = 0x789C
)

type ZlibLevel int

const (
	ZlibNoCompression      ZlibLevel = flate.NoCompression
	ZlibBestSpeedLevel     ZlibLevel = flate.BestSpeed
	ZlibBestCompression    ZlibLevel = flate.BestCompression
	ZlibDefaultCompression ZlibLevel = flate.DefaultCompression
)

// #endregion

// #region Canvas

type CanvasPixformat int32

const (
	CPARGB4444  CanvasPixformat = 1
	CPARGB8888  CanvasPixformat = 2
	CPARGB4444T CanvasPixformat = 3
	CPARGB1555  CanvasPixformat = 257
	CPRGB565    CanvasPixformat = 513
	CPRGB565T   CanvasPixformat = 517
	CPDXT3      CanvasPixformat = 1026
	CPDXT5      CanvasPixformat = 2050
)

type CanvasLevelMap int32

const (
	CLALL16      CanvasLevelMap = 1
	CLALL32      CanvasLevelMap = 2
	CLALL56      CanvasLevelMap = 513
	CLUSE32OVER1 CanvasLevelMap = 65538
	CLUSE32OVER2 CanvasLevelMap = 131074
	CLUSE56OVER1 CanvasLevelMap = 66049
	CLUSE56OVER2 CanvasLevelMap = 131585
)

// #endregion

// #region Sound

type SoundType int8

const (
	SoundMP3 SoundType = iota + 1
	SoundWave
)

const (
	WaveFormatExSize         int32 = 18
	MpegLayer3WaveFormatSize int32 = 30
)

// #endregion
