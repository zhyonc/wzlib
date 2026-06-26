package wzlib

// lut.img.
type wzRawData struct {
	IWzNode

	isBack bool
	data   []byte
}

func NewWzRawData(parent IWzNode) IWzRawData {
	n := new(wzRawData)
	n.IWzNode = NewWzNode(n, NodeRawData, parent)
	return n
}

// DeSerialize implements [IWzRawData].
func (n *wzRawData) DeSerialize(stream IWzStream) {
	stream.Skip(1)
	if stream.Decode1() != 0 {
		// older version only skip 1
		stream.Back(1)
		n.isBack = true
	}
	dataLen := stream.DecodeVT4()
	n.data = stream.DecodeBuffer(int64(dataLen))
}

// Serialize implements [IWzRawData].
func (n *wzRawData) Serialize(archive IWzArchive) {
	if n.isBack {
		// older version only skip 1
		archive.Encode1(0)
	} else {
		archive.Encode2(0)
	}
	archive.EncodeVT4(int32(len(n.data)))
	archive.EncodeBuffer(n.data)
}

// GetSelfNode implements [IWzRawData].
func (n *wzRawData) GetSelfNode() IWzNode {
	return n
}

// GetData implements [IWzRawData].
func (n *wzRawData) GetData() []byte {
	return n.data
}

// SetData implements [IWzRawData].
func (n *wzRawData) SetData(data []byte) {
	n.data = data
}
