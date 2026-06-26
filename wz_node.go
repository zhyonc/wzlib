package wzlib

import (
	"fmt"
	"slices"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
)

type wzNode struct {
	delegate IWzSerialize
	nType    NodeType
	flag     FlagType
	name     string
	nameRef  int32
	size     int32
	checkSum int32
	offset   int32
	root     IWzRootNode
	parent   IWzNode
	childs   *orderedmap.OrderedMap[string, IWzNode]
}

func NewWzNode(delegate IWzSerialize, nType NodeType, parent IWzNode) IWzNode {
	n := &wzNode{
		delegate: delegate,
		nType:    nType,
		parent:   parent,
		childs:   orderedmap.NewOrderedMap[string, IWzNode](),
	}
	return n
}

// DeSerialize implements [IWzNode].
func (n *wzNode) DeSerialize(stream IWzStream) {
	n.delegate.DeSerialize(stream)
}

// Serialize implements [IWzNode].
func (n *wzNode) Serialize(archive IWzArchive) {
	n.delegate.Serialize(archive)
}

// GetSelfNode implements [IWzNode].
func (n *wzNode) GetSelfNode() IWzNode {
	return n.delegate.GetSelfNode()
}

// GetType implements [IWzNode].
func (n *wzNode) GetType() NodeType {
	return n.nType
}

// SetType implements [IWzNode].
func (n *wzNode) SetType(nType NodeType) {
	n.nType = nType
}

// GetFlag implements [IWzNode].
func (n *wzNode) GetFlag() FlagType {
	return n.flag
}

// SetFlag implements [IWzNode].
func (n *wzNode) SetFlag(flag FlagType) {
	n.flag = flag
}

// GetName implements [IWzNode].
func (n *wzNode) GetName() string {
	return n.name
}

// SetName implements [IWzNode].
func (n *wzNode) SetName(name string) {
	n.name = name
}

// GetNameRef implements [IWzNode].
func (n *wzNode) GetNameRef() int32 {
	return n.nameRef
}

// SetNameRef implements [IWzNode].
func (n *wzNode) SetNameRef(offset int32) {
	n.nameRef = offset
}

// GetSize implements [IWzNode].
func (n *wzNode) GetSize() int32 {
	return n.size
}

// SetSize implements [IWzNode].
func (n *wzNode) SetSize(size int32) {
	n.size = size
}

// GetCheckSum implements [IWzNode].
func (n *wzNode) GetCheckSum() int32 {
	return n.checkSum
}

// SetCheckSum implements [IWzNode].
func (n *wzNode) SetCheckSum(sum int32) {
	n.checkSum = sum
}

// GetOffset implements [IWzNode].
func (n *wzNode) GetOffset() int32 {
	return n.offset
}

// SetOffset implements [IWzNode].
func (n *wzNode) SetOffset(offset int32) {
	n.offset = offset
}

// GetRootNode implements [IWzNode].
func (n *wzNode) GetRootNode() IWzRootNode {
	return n.root
}

// SetRootNode implements [IWzNode].
func (n *wzNode) SetRootNode(root IWzRootNode) {
	n.root = root
}

// GetMetaNode implements [IWzNode].
func (n *wzNode) GetMetaNode() IWzNode {
	node := IWzNode(n)
	for {
		parent := node.GetParent()
		if parent == nil {
			return node
		}
		nodeType := parent.GetType()
		if nodeType == NodeDirectory ||
			nodeType == NodeImage ||
			nodeType == NodeImageRef {
			return parent
		}
		node = parent
	}
}

// GetDataOffset implements [IWzNode].
func (n *wzNode) GetDataOffset() int32 {
	return n.GetMetaNode().GetOffset()
}

// Traversal implements [IWzNode].
func (n *wzNode) Traversal(paths []string, stream IWzStream) (IWzNode, error) {
	if len(paths) == 0 {
		return n.GetSelfNode(), nil
	}
	name := paths[0]
	child, ok := n.childs.Get(name)
	if !ok && n.GetChildsLen() == 0 {
		stream.SetOffset(int64(n.GetOffset()))
		WzDispatchDeSerialize(n, stream)
		child, ok = n.childs.Get(name)
	}
	if !ok {
		return nil, fmt.Errorf("child %s not found under %s", name, n.name)
	}
	return child.Traversal(paths[1:], stream)
}

// TraverslChild implements [IWzNode].
func (n *wzNode) TraverslChild(paths []string) (IWzNode, error) {
	if len(paths) == 0 {
		return n.GetSelfNode(), nil
	}
	name := paths[0]
	child, err := n.GetChild(name)
	if err != nil {
		return nil, err
	}
	return child.TraverslChild(paths[1:])
}

// GetParent implements [IWzNode].
func (n *wzNode) GetParent() IWzNode {
	return n.parent
}

// GetChildsLen implements [IWzNode].
func (n *wzNode) GetChildsLen() int32 {
	return int32(n.childs.Len())
}

// GetChildNames implements [IWzNode].
func (n *wzNode) GetChildNames() []string {
	return slices.Collect(n.childs.Keys())
}

// GetChilds implements [IWzNode].
func (n *wzNode) GetChilds() []IWzNode {
	return slices.Collect(n.childs.Values())
}

// GetFirstChild implements [IWzNode].
func (n *wzNode) GetFirstChild() IWzNode {
	if n.childs.Len() == 0 {
		return nil
	}
	return n.childs.Front().Value
}

// GetChild implements [IWzNode].
func (n *wzNode) GetChild(name string) (IWzNode, error) {
	node, ok := n.childs.Get(name)
	if !ok {
		return nil, fmt.Errorf("failed to get child name %s", name)
	}
	return node, nil
}

// GetChildByPath implements [IWzNode].
func (n *wzNode) GetChildByPath(nodePath string) (IWzNode, error) {
	paths := strings.Split(nodePath, DefaultNodePathSeparator)
	if len(paths) == 0 {
		return nil, fmt.Errorf("invalid node path %s", nodePath)
	}
	return n.TraverslChild(paths)
}

// AddChild implements [IWzNode].
func (n *wzNode) AddChild(node IWzNode) {
	n.childs.Set(node.GetName(), node)
}
