package wzlib

type wzVector struct {
	IWzNode

	x int32
	y int32
}

func NewWzVector(parent IWzNode) IWzVector {
	n := new(wzVector)
	n.IWzNode = NewWzNode(n, NodeVector2D, parent)
	return n
}

// DeSerialize implements [IWzVector].
func (n *wzVector) DeSerialize(stream IWzStream) {
	n.x = stream.DecodeVT4()
	n.y = stream.DecodeVT4()
}

// Serialize implements [IWzVector].
func (n *wzVector) Serialize(archive IWzArchive) {
	archive.EncodeVT4(n.x)
	archive.EncodeVT4(n.y)
}

// GetSelfNode implements [IWzVector].
func (n *wzVector) GetSelfNode() IWzNode {
	return n
}

// GetX implements [IWzVector].
func (n *wzVector) GetX() int32 {
	return n.x
}

// SetX implements [IWzVector].
func (n *wzVector) SetX(x int32) {
	n.x = x
}

// GetY implements [IWzVector].
func (n *wzVector) GetY() int32 {
	return n.y
}

// SetY implements [IWzVector].
func (n *wzVector) SetY(y int32) {
	n.y = y
}
