package wzlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
)

type WaveFormatEx struct {
	FormatTag      int16 // PCM(0x01) or MP3Ex(0x55)
	Channels       int16
	SamplesPerSec  int32
	AvgBytesPerSec int32
	BlockAlign     int16
	BitsPerSample  int16
	CBSize         int16 // Size of extra information in bytes (PCM=0, MP3Ex=12)
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

func (wave *WaveFormatEx) Extract(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.LittleEndian, wave.FormatTag)
	binary.Write(buffer, binary.LittleEndian, wave.Channels)
	binary.Write(buffer, binary.LittleEndian, wave.SamplesPerSec)
	binary.Write(buffer, binary.LittleEndian, wave.AvgBytesPerSec)
	binary.Write(buffer, binary.LittleEndian, wave.BlockAlign)
	binary.Write(buffer, binary.LittleEndian, wave.BitsPerSample)
	binary.Write(buffer, binary.LittleEndian, wave.CBSize)
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

func (mp3 *MpegLayer3WaveFormat) Extract(buffer *bytes.Buffer) {
	mp3.WaveFormatEx.Extract(buffer)
	binary.Write(buffer, binary.LittleEndian, mp3.ID)
	binary.Write(buffer, binary.LittleEndian, mp3.Flags)
	binary.Write(buffer, binary.LittleEndian, mp3.BlockSize)
	binary.Write(buffer, binary.LittleEndian, mp3.FramesPerBlock)
	binary.Write(buffer, binary.LittleEndian, mp3.CodecDelay)
}

type wzSound struct {
	IWzNode

	unkFlag1     int8
	duration     int32
	soundType    SoundType
	mediaType    []byte
	mediaSubType []byte
	unkFlag2     int8
	unkFlag3     int8
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
	n.unkFlag1 = stream.Decode1()
	soundDataLen := stream.DecodeVT4()
	n.duration = stream.DecodeVT4()
	n.soundType = SoundType(stream.Decode1())
	n.mediaType = stream.DecodeBuffer(16)    // GUID Stream
	n.mediaSubType = stream.DecodeBuffer(16) // GUID Wave(type2) or MPEG1Audio(type1)
	n.unkFlag2 = stream.Decode1()            // 0 or 1
	n.unkFlag3 = stream.Decode1()            // always 1
	n.formatType = stream.DecodeBuffer(16)   // GUID FORMAT_WaveFormatEx(type2)
	if n.soundType == SoundWave {
		formatSize := stream.DecodeVT4()
		switch formatSize {
		case WaveFormatExSize:
			format := new(WaveFormatEx)
			format.DeSerialize(stream)
			n.format = format
		case MpegLayer3WaveFormatSize:
			format := new(MpegLayer3WaveFormat)
			format.DeSerialize(stream)
			n.format = format
		default:
			slog.Error("Unsupported sound format type", "offset", stream.GetOffset())
			return
		}
	}
	n.soundData = stream.DecodeBuffer(int64(soundDataLen))
}

// Serialize implements [IWzSound].
func (n *wzSound) Serialize(archive IWzArchive) {
	archive.Encode1(n.unkFlag1)
	archive.EncodeVT4(int32(len(n.soundData)))
	archive.EncodeVT4(n.duration)
	archive.Encode1(int8(n.soundType))
	archive.EncodeBuffer(n.mediaType)
	archive.EncodeBuffer(n.mediaSubType)
	archive.Encode1(n.unkFlag2)
	archive.Encode1(n.unkFlag3)
	archive.EncodeBuffer(n.formatType)
	if n.soundType == SoundWave {
		switch format := n.format.(type) {
		case *WaveFormatEx:
			archive.EncodeVT4(WaveFormatExSize)
			format.Serialize(archive)
		case *MpegLayer3WaveFormat:
			archive.EncodeVT4(MpegLayer3WaveFormatSize)
			format.Serialize(archive)
		default:
			slog.Error("Unsupported sound format type", "offset", archive.GetOffset())
			return
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

// GetRawData implements [IWzSound].
func (n *wzSound) GetRawData() ([]byte, error) {
	raw := make([]byte, len(n.soundData))
	copy(raw, n.soundData)
	return raw, nil
}

// ExtractAudio implements [IWzSound].
func (n *wzSound) ExtractAudio() ([]byte, error) {
	raw, err := n.GetRawData()
	if err != nil {
		return nil, err
	}
	switch n.soundType {
	case SoundMP3:
		return raw, nil
	case SoundWave:
		buffer := new(bytes.Buffer)
		// RIFF header
		buffer.WriteString("RIFF")
		buffer.Write([]byte{0, 0, 0, 0}) // chunkSize
		buffer.WriteString("WAVE")
		// fmt chunk
		buffer.WriteString("fmt ")
		var fmtSize int32
		switch format := n.format.(type) {
		case *WaveFormatEx:
			fmtSize = WaveFormatExSize
			binary.Write(buffer, binary.LittleEndian, fmtSize)
			format.Extract(buffer)
		case *MpegLayer3WaveFormat:
			fmtSize = 30
			binary.Write(buffer, binary.LittleEndian, fmtSize)
			format.Extract(buffer)
		default:
			return nil, errors.New("unsupported sound format type")
		}
		// data chunk
		buffer.WriteString("data")
		binary.Write(buffer, binary.LittleEndian, uint32(len(raw)))
		buffer.Write(raw)
		// Write chunkSize
		data := buffer.Bytes()
		chunkSize := uint32(len(data) - 8)
		binary.LittleEndian.PutUint32(data[4:8], chunkSize)
		return data, nil
	default:
		return nil, fmt.Errorf("unknown sound type %d", n.soundType)
	}
}

// ExtractBlob implements [IWzSound].
func (n *wzSound) ExtractBlob() ([]byte, error) {
	if n.duration == 1000 && n.format != nil {
		var samplesPerSec int32
		switch format := n.format.(type) {
		case *WaveFormatEx:
			samplesPerSec = format.SamplesPerSec
		case *MpegLayer3WaveFormat:
			samplesPerSec = format.WaveFormatEx.SamplesPerSec
		default:
			return nil, errors.New("unsupported sound format type")
		}
		if samplesPerSec != int32(len(n.soundData)) {
			return nil, errors.New("it's not blob type")
		}
	}
	raw, err := n.GetRawData()
	if err != nil {
		return nil, err
	}
	return raw, nil
}
