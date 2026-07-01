package wzlib

type wzLua struct {
	IWzNode

	unkFlag int8
	data    []byte
}

func NewWzLua(parent IWzNode) IWzLua {
	n := new(wzLua)
	n.IWzNode = NewWzNode(n, NodeLua, parent)
	return n
}

// DeSerialize implements [IWzLua].
func (n *wzLua) DeSerialize(stream IWzStream) {
	n.unkFlag = stream.Decode1()
	dataLen := stream.DecodeVT4()
	n.data = stream.DecryptBuffer(int64(dataLen))
}

// Serialize implements [IWzLua].
func (n *wzLua) Serialize(archive IWzArchive) {
	archive.Encode1(n.unkFlag)
	archive.EncodeVT4(int32(len(n.data)))
	archive.EncryptBuffer(n.data)
}

// GetSelfNode implements [IWzLua].
func (n *wzLua) GetSelfNode() IWzNode {
	return n
}

// GetData implements [IWzLua].
func (n *wzLua) GetData() []byte {
	return n.data
}

// SetData implements [IWzLua].
func (n *wzLua) SetData(data []byte) {
	n.data = data
}
