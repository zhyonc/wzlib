package wzlib

import "log/slog"

type wzConvex struct {
	IWzNode

	footholds []IWzVector
}

func NewWzConvex(parent IWzNode) IWzConvex {
	n := new(wzConvex)
	n.IWzNode = NewWzNode(n, NodeConvex2D, parent)
	return n
}

// DeSerialize implements [IWzConvex].
func (n *wzConvex) DeSerialize(stream IWzStream) {
	nodeLen := stream.DecodeVT4()
	for range nodeLen {
		node := WzDispatchDeSerialize(n, stream)
		foothold, _ := node.(IWzVector)
		n.footholds = append(n.footholds, foothold)
	}
}

// Serialize implements [IWzConvex].
func (n *wzConvex) Serialize(archive IWzArchive) {
	archive.EncodeVT4(int32(len(n.footholds)))
	for _, foothold := range n.footholds {
		WzDispatchSerialize(foothold, archive)
	}
}

// GetSelfNode implements [IWzConvex].
func (n *wzConvex) GetSelfNode() IWzNode {
	return n
}

// GetFoothold implements [IWzConvex].
func (n *wzConvex) GetFoothold(index int) IWzVector {
	if index < 0 || index > len(n.footholds)-1 {
		slog.Error("Failed to get foothold", "index", index)
		return nil
	}
	return n.footholds[index]
}

// SetFoothold implements [IWzConvex].
func (n *wzConvex) SetFoothold(index int, foothold IWzVector) {
	if index < 0 || index > len(n.footholds)-1 {
		slog.Error("Failed to set foothold", "index", index)
		return
	}
	n.footholds[index] = foothold
}

// GetFootholds implements [IWzConvex].
func (n *wzConvex) GetFootholds() []IWzVector {
	return n.footholds
}

// SetFootholds implements [IWzConvex].
func (n *wzConvex) SetFootholds(footholds []IWzVector) {
	n.footholds = footholds
}
