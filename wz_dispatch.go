package wzlib

import (
	"log/slog"
)

func WzDispatchDeSerialize(parent IWzNode, stream IWzStream) IWzNode {
	var nameRef int32
	var name string
	flag := FlagType(stream.Decode1())
	switch flag {
	case FlagPropName:
		name = stream.DecryptVTStr()
	case FlagPropNameRef:
		nameRef = stream.Decode4()
		name = stream.DecryptVTStrRef(nameRef, parent.GetDataOffset())
	case FlagVTStr, FlagVTStrRef:
		slog.Error("Unsupported flag", "flag", flag)
		return nil
	default:
		slog.Error("Unknown flag", "flag", flag, "offset", stream.GetOffset())
		return nil
	}
	var node IWzNode
	tag := NodeNameTag(name)
	switch tag {
	case NodeNameProperty:
		node = NewWzProperty(parent)
	case NodeNameCanvas:
		node = NewWzCanvas(parent)
	case NodeNameVector2D:
		node = NewWzVector(parent)
	case NodeNameConvex2D:
		node = NewWzConvex(parent)
	case NodeNameSoundDX8:
		node = NewWzSound(parent)
	case NodeNameUOL:
		node = NewWzUOL(parent)
	case NodeNameRawData:
		node = NewWzRawData(parent)
	case NodeNameScript:
		node = NewWzLua(parent)
	default:
		slog.Error("Unknown node name", "tag", tag, "offset", stream.GetOffset())
		return nil
	}
	node.SetFlag(flag)
	node.SetNameRef(nameRef)
	node.SetName(name)
	node.DeSerialize(stream)
	return node
}

func WzDispatchSerialize(node IWzNode, archive IWzArchive) {
	flag := node.GetFlag()
	archive.Encode1(int8(flag))
	switch flag {
	case FlagPropName:
		archive.EncryptVTStr(node.GetName())
	case FlagPropNameRef:
		archive.EncryptVTStrRef(node.GetNameRef())
	case FlagVTStr, FlagVTStrRef:
		slog.Error("Unsupported flag", "flag", flag)
		return
	default:
		slog.Error("Unknown flag", "flag", flag, "offset", archive.GetOffset())
		return
	}
	node.Serialize(archive)
}
