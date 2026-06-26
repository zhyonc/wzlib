package wzlib

import (
	"fmt"
	"log/slog"
)

type wzPropertyItem struct {
	IWzNode

	valueType VARENUM
	valueFlag FlagType
	valueRef  int32
	value     any
}

func NewWzPropertyItem(parent IWzNode) IWzPropertyItem {
	pi := new(wzPropertyItem)
	pi.IWzNode = NewWzNode(pi, NodePropertyItem, parent)
	return pi
}

// DeSerialize implements [IWzPropertyItem].
func (pi *wzPropertyItem) DeSerialize(stream IWzStream) {
	flag := FlagType(stream.Decode1())
	pi.SetFlag(flag)
	switch flag {
	case FlagVTStr:
		pi.SetName(stream.DecryptVTStr())
	case FlagVTStrRef:
		ref := stream.Decode4()
		name := stream.DecryptVTStrRef(ref, pi.GetDataOffset())
		pi.SetNameRef(ref)
		pi.SetName(name)
	case FlagPropName, FlagPropNameRef:
		slog.Error("Unsupported flag", "flag", flag)
		return
	default:
		slog.Error("Failed to decrypt item name", "flag", flag, "offset", stream.GetOffset())
		return
	}
	pi.valueType = VARENUM(stream.DecodeVT4())
	switch pi.valueType { //nolint:exhaustive // not all VARENUM cases are needed
	case VTEmpty, VTNull:
		pi.value = nil
	case VTI2, VTUI2, VTBool:
		pi.value = stream.Decode2()
	case VTInt, VTUInt:
		pi.value = stream.Decode4()
	case VTI4, VTUI4:
		pi.value = stream.DecodeVT4()
	case VTR4:
		pi.value = stream.DecodeVT4f()
	case VTI8, VTUI8:
		pi.value = stream.DecodeVT8()
	case VTR8:
		pi.value = stream.Decode8f()
	case VTBStr:
		flag = FlagType(stream.Decode1())
		pi.valueFlag = flag
		switch flag {
		case FlagVTStr:
			pi.value = stream.DecryptVTStr()
		case FlagVTStrRef:
			ref := stream.Decode4()
			str := stream.DecryptVTStrRef(ref, pi.GetDataOffset())
			pi.valueRef = ref
			pi.value = str
		case FlagPropName, FlagPropNameRef:
			slog.Error("Unsupported flag", "flag", flag)
			return
		default:
			slog.Error("Failed to decrypt VT_BSTR", "flag", flag, "offset", stream.GetOffset())
			return
		}
	case VTDispatch:
		pi.SetSize(stream.Decode4())
		pi.value = WzDispatchDeSerialize(pi, stream)
	default:
		slog.Error("Unknown value type", "valueType", pi.valueType, "offset", stream.GetOffset())
		return
	}
}

// Serialize implements [IWzPropertyItem].
func (pi *wzPropertyItem) Serialize(archive IWzArchive) {
	flag := pi.GetFlag()
	archive.Encode1(int8(flag))
	switch flag {
	case FlagVTStr:
		archive.EncryptVTStr(pi.GetName())
	case FlagVTStrRef:
		archive.EncryptVTStrRef(pi.GetNameRef())
	case FlagPropName, FlagPropNameRef:
		slog.Error("Unsupported flag", "flag", flag)
		return
	}
	archive.EncodeVT4(int32(pi.GetValueType()))
	switch pi.valueType { //nolint:exhaustive // not all VARENUM cases are needed
	case VTEmpty, VTNull:
	case VTI2, VTUI2, VTBool:
		v, _ := pi.value.(int16)
		archive.Encode2(v)
	case VTInt, VTUInt:
		v, _ := pi.value.(int32)
		archive.Encode4(v)
	case VTI4, VTUI4:
		v, _ := pi.value.(int32)
		archive.EncodeVT4(v)
	case VTR4:
		v, _ := pi.value.(float32)
		archive.EncodeVT4f(v)
	case VTI8, VTUI8:
		v, _ := pi.value.(int64)
		archive.EncodeVT8(v)
	case VTR8:
		v, _ := pi.value.(float64)
		archive.Encode8f(v)
	case VTBStr:
		archive.Encode1(int8(pi.valueFlag))
		switch pi.valueFlag {
		case FlagVTStr:
			s, _ := pi.value.(string)
			archive.EncryptVTStr(s)
		case FlagVTStrRef:
			archive.EncryptVTStrRef(pi.valueRef)
		case FlagPropName, FlagPropNameRef:
			slog.Error("Unsupported flag", "flag", flag)
			return
		}
	case VTDispatch:
		archive.Encode4(pi.GetSize())
		node, _ := pi.value.(IWzNode)
		WzDispatchSerialize(node, archive)
	default:
		slog.Error("Unknown value type", "valueType", pi.valueType, "offset", archive.GetOffset())
		return
	}
}

// GetSelfNode implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetSelfNode() IWzNode {
	return pi
}

// GetValueType implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetValueType() VARENUM {
	return pi.valueType
}

// SetValueType implements [IWzPropertyItem].
func (pi *wzPropertyItem) SetValueType(vt VARENUM) {
	pi.valueType = vt
}

// GetValue implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetValue() any {
	return pi.value
}

// SetValue implements [IWzPropertyItem].
func (pi *wzPropertyItem) SetValue(value any) {
	pi.value = value
}

// SetVTStr implements [IWzPropertyItem].
func (pi *wzPropertyItem) SetVTStr(str string) {
	pi.valueFlag = FlagVTStr
	pi.value = str
}

// SetVTStrRef implements [IWzPropertyItem].
func (pi *wzPropertyItem) SetVTStrRef(strRef int32) {
	pi.valueFlag = FlagVTStrRef
	pi.valueRef = strRef
}

func getValue[T any](pi IWzPropertyItem) (T, error) {
	v, ok := pi.GetValue().(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("property item %s: failed to assert %T", pi.GetName(), zero)
	}
	return v, nil
}

// GetInt16 implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetInt16() (int16, error) {
	return getValue[int16](pi)
}

// GetInt32 implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetInt32() (int32, error) {
	return getValue[int32](pi)
}

// GetInt64 implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetInt64() (int64, error) {
	return getValue[int64](pi)
}

// GetFloat32 implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetFloat32() (float32, error) {
	return getValue[float32](pi)
}

// GetFloat64 implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetFloat64() (float64, error) {
	return getValue[float64](pi)
}

// GetString implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetString() (string, error) {
	return getValue[string](pi)
}

// GetCanvas implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetCanvas() (IWzCanvas, error) {
	return getValue[IWzCanvas](pi)
}

// GetVector implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetVector() (IWzVector, error) {
	return getValue[IWzVector](pi)
}

// GetConvex implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetConvex() (IWzConvex, error) {
	return getValue[IWzConvex](pi)
}

// GetSound implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetSound() (IWzSound, error) {
	return getValue[IWzSound](pi)
}

// GetUOL implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetUOL() (IWzUOL, error) {
	return getValue[IWzUOL](pi)
}

// GetRawData implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetRawData() (IWzRawData, error) {
	return getValue[IWzRawData](pi)
}

// GetScript implements [IWzPropertyItem].
func (pi *wzPropertyItem) GetScript() (IWzLua, error) {
	return getValue[IWzLua](pi)
}
