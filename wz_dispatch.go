package wzlib

import (
	"log/slog"
	"strings"
)

func WzDispatchDeSerialize(parent IWzNode, stream IWzStream) IWzNode {
	if strings.Contains(parent.GetName(), ".lua") {
		// GMS232 Etc.wz/Script/*.lua is NodeImage type
		node := NewWzLua(parent)
		node.DeSerialize(stream)
		parent.AddChild(node)
		return nil
	}
	var name string
	var nameRef int32
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
	default:
		slog.Error("Unknown node name", "tag", tag, "offset", stream.GetOffset())
		return nil
	}
	node.SetFlag(flag)
	node.SetName(name)
	node.SetNameRef(nameRef)
	node.DeSerialize(stream)
	return node
}

func WzDispatchSerialize(node IWzNode, archive IWzArchive) {
	parentName := node.GetParent().GetName()
	if strings.Contains(parentName, ".lua") {
		luaNode := node.GetParent().GetFirstChild()
		if luaNode == nil {
			slog.Error("Miss lua node", "parentName", parentName)
			return
		}
		luaNode.Serialize(archive)
		return
	}
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
