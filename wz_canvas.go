package wzlib

type wzCanvas struct {
	IWzNode

	prop        IWzProperty
	width       int32
	height      int32
	pixelFormat CanvasPixformat
	magLevel    int8
	levelMap    CanvasLevelMap
	data        []byte
}

func NewWzCanvas(parent IWzNode) IWzCanvas {
	n := new(wzCanvas)
	n.IWzNode = NewWzNode(n, NodeCanvas, parent)
	return n
}

// DeSerialize implements [IWzCanvas].
func (n *wzCanvas) DeSerialize(stream IWzStream) {
	stream.Skip(1)
	hasProperty := stream.DecodeBool()
	if hasProperty {
		n.prop = NewWzProperty(n.GetParent())
		n.prop.SetName(string(NodeNameProperty))
		n.prop.DeSerialize(stream)
	}
	n.width = stream.DecodeVT4()
	n.height = stream.DecodeVT4()
	n.pixelFormat = CanvasPixformat(stream.DecodeVT4())
	n.magLevel = stream.Decode1()
	n.levelMap = CanvasLevelMap(stream.Decode4())
	dataLen := stream.Decode4()
	n.data = stream.DecodeBuffer(int64(dataLen)) // zlib with deflate
}

// Serialize implements [IWzCanvas].
func (n *wzCanvas) Serialize(archive IWzArchive) {
	archive.Encode1(0)
	archive.EncodeBool(n.prop != nil)
	if n.prop != nil {
		n.prop.Serialize(archive)
	}
	archive.EncodeVT4(n.width)
	archive.EncodeVT4(n.height)
	archive.EncodeVT4(int32(n.pixelFormat))
	archive.Encode1(n.magLevel)
	archive.Encode4(int32(n.levelMap))
	archive.Encode4(int32(len(n.data)))
	archive.EncodeBuffer(n.data)
}

// GetSelfNode implements [IWzCanvas].
func (n *wzCanvas) GetSelfNode() IWzNode {
	return n
}

// GetProperty implements [IWzCanvas].
func (n *wzCanvas) GetProperty() IWzProperty {
	return n.prop
}

// SetProperty implements [IWzCanvas].
func (n *wzCanvas) SetProperty(prop IWzProperty) {
	n.prop = prop
}

// GetWidth implements [IWzCanvas].
func (n *wzCanvas) GetWidth() int32 {
	return n.width
}

// SetWidth implements [IWzCanvas].
func (n *wzCanvas) SetWidth(width int32) {
	n.width = width
}

// GetHeight implements [IWzCanvas].
func (n *wzCanvas) GetHeight() int32 {
	return n.height
}

// SetHeight implements [IWzCanvas].
func (n *wzCanvas) SetHeight(height int32) {
	n.height = height
}

// GetPixelFormat implements [IWzCanvas].
func (n *wzCanvas) GetPixelFormat() CanvasPixformat {
	return n.pixelFormat
}

// SetPixelFormat implements [IWzCanvas].
func (n *wzCanvas) SetPixelFormat(cp CanvasPixformat) {
	n.pixelFormat = cp
}

// GetMagLevel implements [IWzCanvas].
func (n *wzCanvas) GetMagLevel() int8 {
	return n.magLevel
}

// SetMagLevel implements [IWzCanvas].
func (n *wzCanvas) SetMagLevel(magLevel int8) {
	n.magLevel = magLevel
}

// GetData implements [IWzCanvas].
func (n *wzCanvas) GetData() []byte {
	return n.data
}

// SetData implements [IWzCanvas].
func (n *wzCanvas) SetData(data []byte) {
	n.data = data
}
