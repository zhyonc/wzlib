package wzlib

import (
	"errors"
)

type wzDirectory struct {
	IWzNode
}

func NewWzDirectory(root IWzRootNode, parent IWzNode) IWzDirectory {
	dir := new(wzDirectory)
	dir.IWzNode = NewWzNode(dir, NodeDirectory, parent)
	dir.SetRootNode(root)
	return dir
}

// DeSerialize implements [IWzDirectory].
func (dir *wzDirectory) DeSerialize(stream IWzStream) {
	dir.SetName(stream.DecryptVTStr())
	dir.SetSize(stream.DecodeVT4())
	dir.SetCheckSum(stream.DecodeVT4())
	dir.SetOffset(stream.DecryptDataOffset(dir.GetRootNode().GetHeaderSize(), dir.GetRootNode().GetVersion()))
}

// Serialize implements [IWzDirectory].
func (dir *wzDirectory) Serialize(archive IWzArchive) {
	archive.Encode1(int8(dir.GetType()))
	archive.EncryptVTStr(dir.GetName())
	archive.EncodeVT4(dir.GetSize())
	archive.EncodeVT4(dir.GetCheckSum())
	archive.EncryptDataOffset(dir.GetOffset(), dir.GetRootNode().GetHeaderSize(), dir.GetRootNode().GetVersion())
}

// GetSelfNode implements [IWzDirectory].
func (dir *wzDirectory) GetSelfNode() IWzNode {
	return dir
}

// GetDirectory implements [IWzDirectory].
func (dir *wzDirectory) GetDirectory(nodePath string) (IWzDirectory, error) {
	child, err := dir.GetChildByPath(nodePath)
	if err != nil {
		return nil, err
	}
	nextDir, ok := child.(IWzDirectory)
	if !ok {
		return nil, errors.New("failed to assert IWzDirectory")
	}
	return nextDir, nil
}

// GetImage implements [IWzDirectory].
func (dir *wzDirectory) GetImage(nodePath string) (IWzImage, error) {
	child, err := dir.GetChildByPath(nodePath)
	if err != nil {
		return nil, err
	}
	img, ok := child.(IWzImage)
	if !ok {
		return nil, errors.New("failed to assert IWzImage")
	}
	if img.GetChildsLen() == 0 {
		dir.GetRootNode().GetStream().SetOffset(int64(img.GetOffset()))
		WzDispatchDeSerialize(img, dir.GetRootNode().GetStream())
	}
	return img, nil
}
