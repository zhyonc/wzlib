package wzlib

type WaveFormatEx struct {
	FormatTag      int16 // MPEGLAYER3(0x55) or PCM(0x01)
	Channels       int16
	SamplesPerSec  int32
	AvgBytesPerSec int32
	BlockAlign     int16
	BitsPerSample  int16
	CBSize         int16 // The count in bytes of the size of extra information
}

func (wave *WaveFormatEx) DeSerialize(stream IWzStream) {
	wave.FormatTag = stream.Decode2()
	wave.Channels = stream.Decode2()
	wave.SamplesPerSec = stream.Decode4()
	wave.AvgBytesPerSec = stream.Decode4()
	wave.BlockAlign = stream.Decode2()
	wave.BitsPerSample = stream.Decode2()
	wave.CBSize = stream.Decode2()
}

func (wave *WaveFormatEx) Serialize(archive IWzArchive) {
	archive.Encode2(wave.FormatTag)
	archive.Encode2(wave.Channels)
	archive.Encode4(wave.SamplesPerSec)
	archive.Encode4(wave.AvgBytesPerSec)
	archive.Encode2(wave.BlockAlign)
	archive.Encode2(wave.BitsPerSample)
	archive.Encode2(wave.CBSize)
}

type MpegLayer3WaveFormat struct {
	WaveFormatEx

	ID             int16
	Flags          int32
	BlockSize      int16
	FramesPerBlock int16
	CodecDelay     int16
}

func (mp3 *MpegLayer3WaveFormat) DeSerialize(stream IWzStream) {
	mp3.WaveFormatEx.DeSerialize(stream)
	mp3.ID = stream.Decode2()
	mp3.Flags = stream.Decode4()
	mp3.BlockSize = stream.Decode2()
	mp3.FramesPerBlock = stream.Decode2()
	mp3.CodecDelay = stream.Decode2()
}

func (mp3 *MpegLayer3WaveFormat) Serialize(archive IWzArchive) {
	mp3.WaveFormatEx.Serialize(archive)
	archive.Encode2(mp3.ID)
	archive.Encode4(mp3.Flags)
	archive.Encode2(mp3.BlockSize)
	archive.Encode2(mp3.FramesPerBlock)
	archive.Encode2(mp3.CodecDelay)
}

type wzSound struct {
	IWzNode

	duration     int32
	soundType    SoundType
	mediaType    []byte
	mediaSubType []byte
	unkFlag1     int8
	unkFlag2     int8
	formatType   []byte
	format       any
	soundData    []byte
}

func NewWzSound(parent IWzNode) IWzSound {
	n := new(wzSound)
	n.IWzNode = NewWzNode(n, NodeSoundDX8, parent)
	return n
}

// DeSerialize implements [IWzSound].
func (n *wzSound) DeSerialize(stream IWzStream) {
	stream.Skip(1)
	soundDataLen := stream.DecodeVT4()
	n.duration = stream.DecodeVT4()
	n.soundType = SoundType(stream.Decode1())
	n.mediaType = stream.DecodeBuffer(16)    // GUID Stream
	n.mediaSubType = stream.DecodeBuffer(16) // GUID Wave(type2) or MPEG1Audio(type1)
	n.unkFlag1 = stream.Decode1()            // 0 or 1
	n.unkFlag2 = stream.Decode1()            // always 1
	n.formatType = stream.DecodeBuffer(16)   // GUID FORMAT_WaveFormatEx(type2)
	if n.soundType == SoundPCM {
		formatSize := stream.DecodeVT4()
		switch formatSize {
		case WaveFormatExSize:
			wave := new(WaveFormatEx)
			wave.DeSerialize(stream)
			n.format = wave
		case MpegLayer3WaveFormatSize:
			mp3 := new(MpegLayer3WaveFormat)
			mp3.DeSerialize(stream)
			n.format = mp3
		}
	}
	n.soundData = stream.DecodeBuffer(int64(soundDataLen))
}

// Serialize implements [IWzSound].
func (n *wzSound) Serialize(archive IWzArchive) {
	archive.Encode1(0)
	archive.EncodeVT4(int32(len(n.soundData)))
	archive.EncodeVT4(n.duration)
	archive.Encode1(int8(n.soundType))
	archive.EncodeBuffer(n.mediaType)
	archive.EncodeBuffer(n.mediaSubType)
	archive.Encode1(n.unkFlag1)
	archive.Encode1(n.unkFlag2)
	archive.EncodeBuffer(n.formatType)
	if n.soundType == SoundPCM {
		switch format := n.format.(type) {
		case *WaveFormatEx:
			archive.EncodeVT4(WaveFormatExSize)
			format.Serialize(archive)
		case *MpegLayer3WaveFormat:
			archive.EncodeVT4(MpegLayer3WaveFormatSize)
			format.Serialize(archive)
		}
	}
	archive.EncodeBuffer(n.soundData)
}

// GetSelfNode implements [IWzSound].
func (n *wzSound) GetSelfNode() IWzNode {
	return n
}

// GetDuration implements [IWzSound].
func (n *wzSound) GetDuration() int32 {
	return n.duration
}

// SetDuration implements [IWzSound].
func (n *wzSound) SetDuration(duration int32) {
	n.duration = duration
}

// GetSoundType implements [IWzSound].
func (n *wzSound) GetSoundType() SoundType {
	return n.soundType
}

// SetSoundType implements [IWzSound].
func (n *wzSound) SetSoundType(soundType SoundType) {
	n.soundType = soundType
}

// GetMediaType implements [IWzSound].
func (n *wzSound) GetMediaType() []byte {
	return n.mediaType
}

// SetMediaType implements [IWzSound].
func (n *wzSound) SetMediaType(mediaType []byte) {
	n.mediaType = mediaType
}

// GetMediaSubType implements [IWzSound].
func (n *wzSound) GetMediaSubType() []byte {
	return n.mediaSubType
}

// SetMediaSubType implements [IWzSound].
func (n *wzSound) SetMediaSubType(mediaSubType []byte) {
	n.mediaSubType = mediaSubType
}

// GetFormatType implements [IWzSound].
func (n *wzSound) GetFormatType() []byte {
	return n.formatType
}

// SetFormatType implements [IWzSound].
func (n *wzSound) SetFormatType(formatType []byte) {
	n.formatType = formatType
}

// GetFormat implements [IWzSound].
func (n *wzSound) GetFormat() any {
	return n.format
}

// SetFormat implements [IWzSound].
func (n *wzSound) SetFormat(format any) {
	n.format = format
}

// GetData implements [IWzSound].
func (n *wzSound) GetData() []byte {
	return n.soundData
}

// SetData implements [IWzSound].
func (n *wzSound) SetData(data []byte) {
	n.soundData = data
}
