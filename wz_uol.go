package wzlib

import "log/slog"

type wzUOL struct {
	IWzNode

	valueflag FlagType
	valueRef  int32
	value     string
}

func NewWzUOL(parent IWzNode) IWzUOL {
	n := new(wzUOL)
	n.IWzNode = NewWzNode(n, NodeUOL, parent)
	return n
}

// DeSerialize implements [IWzUOL].
func (n *wzUOL) DeSerialize(stream IWzStream) {
	stream.Skip(1)
	n.valueflag = FlagType(stream.Decode1())
	switch n.valueflag {
	case FlagVTStr:
		n.value = stream.DecryptVTStr()
	case FlagVTStrRef:
		n.valueRef = stream.Decode4()
		n.value = stream.DecryptVTStrRef(n.valueRef, n.GetDataOffset())
	case FlagPropName, FlagPropNameRef:
		slog.Error("Unsupported flag", "flag", n.valueflag)
		return
	}
}

// Serialize implements [IWzUOL].
func (n *wzUOL) Serialize(archive IWzArchive) {
	archive.Encode1(0)
	archive.Encode1(int8(n.valueflag))
	switch n.valueflag {
	case FlagVTStr:
		archive.EncryptVTStr(n.value)
	case FlagVTStrRef:
		archive.EncryptVTStrRef(n.valueRef)
	case FlagPropName, FlagPropNameRef:
		slog.Error("Unsupported flag", "flag", n.valueflag)
		return
	}
}

// GetSelfNode implements [IWzUOL].
func (n *wzUOL) GetSelfNode() IWzNode {
	return n
}

// GetUOL implements [IWzUOL].
func (n *wzUOL) GetUOL() string {
	return n.value
}

// SetUOL implements [IWzUOL].
func (n *wzUOL) SetUOL(uol string) {
	n.value = uol
}
