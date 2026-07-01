package wzlib

import (
	"log/slog"
)

type wzProperty struct {
	IWzNode
}

func NewWzProperty(parent IWzNode) IWzProperty {
	p := new(wzProperty)
	p.IWzNode = NewWzNode(p, NodeProperty, parent)
	return p
}

// DeSerialize implements [IWzProperty].
func (p *wzProperty) DeSerialize(stream IWzStream) {
	parent := p.GetParent()
	stream.Skip(2)
	itemLen := stream.DecodeVT4()
	for range itemLen {
		item := NewWzPropertyItem(parent)
		item.DeSerialize(stream)
		parent.AddChild(item)
	}
}

// Serialize implements [IWzProperty].
func (p *wzProperty) Serialize(archive IWzArchive) {
	childs := p.GetParent().GetChilds()
	archive.Encode2(0)
	archive.EncodeVT4(int32(len(childs)))
	for _, child := range childs {
		item, ok := child.(IWzPropertyItem)
		if !ok {
			slog.Error("Failed to assert IWzPropertyItem", "offset", archive.GetOffset())
			return
		}
		item.Serialize(archive)
	}
}

// GetSelfNode implements [IWzProperty].
func (p *wzProperty) GetSelfNode() IWzNode {
	return p
}
