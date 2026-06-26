package wzlib

import (
	"errors"
	"log/slog"
)

type wzImage struct {
	IWzNode
}

func NewWzImage(root IWzRootNode, parent IWzNode) IWzImage {
	img := new(wzImage)
	img.IWzNode = NewWzNode(img, NodeImage, parent)
	img.SetRootNode(root)
	return img
}

// DeSerialize implements [IWzImage].
func (img *wzImage) DeSerialize(stream IWzStream) {
	nType := img.GetType()
	switch nType { //nolint:exhaustive // not all NodeType cases are needed
	case NodeImage:
		img.SetName(stream.DecryptVTStr())
	case NodeImageRef:
		nameRef := stream.Decode4()
		name := stream.DecryptVTStrRef(nameRef, img.GetRootNode().GetHeaderSize()+1)
		img.SetNameRef(nameRef)
		img.SetName(name)
	default:
		slog.Error("Unknown node type", "type", img.GetType(), "offset", stream.GetOffset())
		return
	}
	img.SetSize(stream.DecodeVT4())
	img.SetCheckSum(stream.DecodeVT4())
	img.SetOffset(stream.DecryptDataOffset(img.GetRootNode().GetHeaderSize(), img.GetRootNode().GetVersion()))
}

// Serialize implements [IWzImage].
func (img *wzImage) Serialize(archive IWzArchive) {
	nType := img.GetType()
	archive.Encode1(int8(nType))
	switch nType { //nolint:exhaustive // not all NodeType cases are needed
	case NodeImage:
		archive.EncryptVTStr(img.GetName())
	case NodeImageRef:
		archive.EncryptVTStrRef(img.GetNameRef())
	default:
		slog.Error("Unknown node type", "type", img.GetType(), "offset", archive.GetOffset())
		return
	}
	archive.EncodeVT4(img.GetSize())
	archive.EncodeVT4(img.GetCheckSum())
	archive.EncryptDataOffset(img.GetOffset(), img.GetRootNode().GetHeaderSize(), img.GetRootNode().GetVersion())
}

// GetSelfNode implements [IWzImage].
func (img *wzImage) GetSelfNode() IWzNode {
	return img
}

// GetPropertyItem implements [IWzImage].
func (img *wzImage) GetPropertyItem(nodePath string) (IWzPropertyItem, error) {
	child, err := img.GetChildByPath(nodePath)
	if err != nil {
		return nil, err
	}
	temp, ok := child.(IWzPropertyItem)
	if !ok {
		return nil, errors.New("failed to assert IWzPropertyItem")
	}
	return temp, nil
}
