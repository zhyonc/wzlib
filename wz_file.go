package wzlib

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v3"
)

type wzFile struct {
	isLoadLazy bool
	aesCipher  IAESCipher
	locale     ILocale
	stream     IWzStream
	ident      string
	size       int64
	headerSize int32
	copyright  string
	version    int16
	nodeMap    *orderedmap.OrderedMap[string, IWzNode]
	imgQueue   []IWzNode
	timeout    time.Duration
	startTime  time.Time
	endTime    time.Time
}

func NewWzFile(setting *Setting) IWzFile {
	f := &wzFile{
		isLoadLazy: setting.IsLazyLoad,
		locale:     NewLocale(setting.LocaleRegion),
		version:    setting.MSVersion,
		nodeMap:    orderedmap.NewOrderedMap[string, IWzNode](),
		timeout:    time.Duration(DefaultTimeoutSec) * time.Second,
	}
	if !setting.DisableAESCipher {
		f.aesCipher = NewAESCipher(setting.MSRegion, setting.IVKey, setting.AESKey)
	}
	return f
}

// GetIdent implements [IWzFile].
func (f *wzFile) GetIdent() string {
	return f.ident
}

// GetSize implements [IWzFile].
func (f *wzFile) GetSize() int64 {
	return f.size
}

// GetHeaderSize implements [IWzFile].
func (f *wzFile) GetHeaderSize() int32 {
	return f.headerSize
}

// GetCopyright implements [IWzFile].
func (f *wzFile) GetCopyright() string {
	return f.copyright
}

// GetVersion implements [IWzFile].
func (f *wzFile) GetVersion() int16 {
	return f.version
}

// GetStream implements [IWzFile].
func (f *wzFile) GetStream() IWzStream {
	return f.stream
}

// GetNode implements [IWzFile].
func (f *wzFile) GetNode(nodePath string) (IWzNode, error) {
	paths := strings.Split(nodePath, NodePathSeparator)
	if len(paths) == 0 {
		return nil, fmt.Errorf("invalid node path %s", nodePath)
	}
	metaNode, ok := f.nodeMap.Get(paths[0])
	if !ok {
		return nil, errors.New("failed to find meta node")
	}
	paths = paths[1:]
	if len(paths) == 0 {
		return metaNode, nil
	}
	node, err := metaNode.Traversal(paths, f.stream)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// GetDirectory implements [IWzFile].
func (f *wzFile) GetDirectory(nodePath string) (IWzDirectory, error) {
	node, err := f.GetNode(nodePath)
	if err != nil {
		return nil, err
	}
	dir, ok := node.(IWzDirectory)
	if !ok {
		return nil, errors.New("failed to assert IWzDirectory")
	}
	return dir, nil
}

// GetImage implements [IWzFile].
func (f *wzFile) GetImage(nodePath string) (IWzImage, error) {
	node, err := f.GetNode(nodePath)
	if err != nil {
		return nil, err
	}
	img, ok := node.(IWzImage)
	if !ok {
		return nil, errors.New("failed to assert IWzImage")
	}
	if img.GetChildsLen() == 0 {
		f.stream.SetOffset(int64(img.GetOffset()))
		WzDispatchDeSerialize(img, f.stream)
	}
	return img, nil
}

// GetPropertyItem implements [IWzFile].
func (f *wzFile) GetPropertyItem(nodePath string) (IWzPropertyItem, error) {
	node, err := f.GetNode(nodePath)
	if err != nil {
		return nil, err
	}
	item, ok := node.(IWzPropertyItem)
	if !ok {
		return nil, errors.New("failed to assert IWzPropertyItem")
	}
	return item, nil
}

// GetMetaNodeLen implements [IWzFile].
func (f *wzFile) GetMetaNodeLen() int {
	return f.nodeMap.Len()
}

// GetMetaNodeKeys implements [IWzFile].
func (f *wzFile) GetMetaNodeKeys() []string {
	return slices.Collect(f.nodeMap.Keys())
}

// Elapsed implements [IWzFile].
func (f *wzFile) Elapsed() time.Duration {
	if f.endTime.IsZero() {
		return time.Since(f.startTime)
	}
	return f.endTime.Sub(f.startTime)
}

// #region Load
// Load implements [IWzFile].
func (f *wzFile) Load(filePath string) error {
	f.startTime = time.Now()
	defer func() {
		f.endTime = time.Now()
	}()
	// New Read Stream
	stream, err := NewWzStream(filePath, f.aesCipher, f.locale)
	if err != nil {
		return err
	}
	if f.isLoadLazy {
		f.stream = stream
	} else {
		defer stream.Close()
	}
	// Timeout check
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	// Header
	err = f.ReadHeader(stream)
	if err != nil {
		return err
	}
	// Meta Node
	err = f.ReadMetaNode(ctx, stream, nil)
	if err != nil {
		return err
	}
	// Data Node
	if f.isLoadLazy {
		return nil
	}
	err = f.ReadDataNode(ctx, stream)
	if err != nil {
		return err
	}
	return nil
}

// ReadHeader implements [IWzFile].
func (f *wzFile) ReadHeader(stream IWzStream) error {
	// Ident
	f.ident = stream.DecodeStr(4)
	if f.ident != DefaultWzFileIdent {
		return fmt.Errorf("invalid ident: expected %s, got %s", DefaultWzFileIdent, f.ident)
	}
	// File Size
	f.size = stream.Decode8()
	if f.size <= 0 {
		return fmt.Errorf("invalid file size: %d", f.size)
	}
	// Header Size
	f.headerSize = stream.Decode4()
	// Copyright
	f.copyright = stream.DecodeNTStr()
	// Version
	switch f.version {
	case 0:
		err := f.PredictVersion(stream)
		if err != nil {
			return err
		}
	case NoHashVersion:
		slog.Debug("The version did not encode 2-byte versionHashSumTruncated")
	default:
		stream.Skip(2)
	}
	return nil
}

// PredictVersion implements [IWzFile].
func (f *wzFile) PredictVersion(stream IWzStream) error {
	var possibleVersions []int16
	versionHashSumTruncated := stream.Decode2()
	if versionHashSumTruncated > 0xFF {
		// versionHashSumTruncated always less than 256
		possibleVersions = []int16{NoHashVersion}
		stream.Back(2)
	} else {
		possibleVersions = GetPossibleVersions(versionHashSumTruncated)
	}
	if len(possibleVersions) == 0 {
		return errors.New("missing version")
	}
	nodesLen := stream.DecodeVT4()
	validNodeCount := 0
	backOffset := stream.GetOffset()
	// Predict version
	for _, version := range possibleVersions {
		f.version = version
		for range nodesLen {
			var node IWzNode
			nType := NodeType(stream.Decode1())
			switch nType { //nolint:exhaustive // not all NodeType cases are needed
			case NodeDirectory:
				node = NewWzDirectory(f, nil)
			case NodeImage:
				node = NewWzImage(f, nil)
			case NodeImageRef:
				node = NewWzImage(f, nil)
				node.SetType(NodeImageRef)
			default:
				return fmt.Errorf("failed to predict version at offset %d", stream.GetOffset())
			}
			node.DeSerialize(stream)
			dataOffset := node.GetDataOffset()
			if dataOffset < 0 {
				validNodeCount = 0
				stream.SetOffset(backOffset)
				break
			}
			validNodeCount++
		}
		if validNodeCount == int(nodesLen) {
			break
		}
	}
	if f.version == 0 {
		return errors.New("failed to predict version")
	}
	slog.Debug("Predicted version", "number", f.version)
	if f.version == NoHashVersion {
		stream.SetOffset(int64(f.headerSize))
	} else {
		stream.SetOffset(int64(f.headerSize) + 2)
	}
	return nil
}

// ReadMetaNode implements [IWzFile].
func (f *wzFile) ReadMetaNode(ctx context.Context, stream IWzStream, parent IWzNode) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Level-order BFS
	dirQueue := make([]IWzNode, 0)
	nodesLen := stream.DecodeVT4()
	for range nodesLen {
		var node IWzNode
		nType := NodeType(stream.Decode1())
		switch nType { //nolint:exhaustive // not all NodeType cases are needed
		case NodeDirectory:
			node = NewWzDirectory(f, parent)
			dirQueue = append(dirQueue, node)
		case NodeImage, NodeImageRef:
			node = NewWzImage(f, parent)
			node.SetType(nType)
			f.imgQueue = append(f.imgQueue, node)
		default:
			return fmt.Errorf("failed to read meta node at offset %d", stream.GetOffset())
		}
		node.DeSerialize(stream)
		if parent == nil {
			f.nodeMap.Set(node.GetName(), node)
		} else {
			parent.AddChild(node)
		}
	}
	for _, dir := range dirQueue {
		err := f.ReadMetaNode(ctx, stream, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadDataNode implements [IWzFile].
func (f *wzFile) ReadDataNode(ctx context.Context, stream IWzStream) error {
	for _, img := range f.imgQueue {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		WzDispatchDeSerialize(img, stream)
	}
	f.imgQueue = f.imgQueue[:0]
	return nil
}

// #endregion

// #region Save
// Save implements [IWzFile].
func (f *wzFile) Save(filePath string, isCreate bool) error {
	f.startTime = time.Now()
	defer func() {
		f.endTime = time.Now()
	}()
	// New Write archive
	archive := NewWzArchive(f.aesCipher, f.locale)
	defer archive.Close()
	// Timeout check
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	// Header
	err := f.WriteHeader(archive)
	if err != nil {
		return err
	}
	if !isCreate {
		// Meta Node
		err = f.WriteMetaNode(ctx, archive, nil)
		if err != nil {
			return err
		}
		// Data Node
		err = f.WriteDataNode(ctx, archive)
		if err != nil {
			return err
		}
	}
	// Update file size
	buf := archive.GetBuffer()
	dataLen := len(buf) - int(f.headerSize)
	binary.LittleEndian.PutUint64(buf[4:12], uint64(dataLen))
	return SaveFile(filePath, buf)
}

// WriteHeader implements [IWzFile].
func (f *wzFile) WriteHeader(archive IWzArchive) error {
	// Ident
	if f.ident == "" {
		f.ident = DefaultWzFileIdent
	}
	archive.EncodeStr(f.ident)
	// File Size
	archive.Encode8(0)
	// Header Size
	if f.headerSize == 0 {
		f.headerSize = int32(DefaultWzFileHeaderSize)
	}
	archive.Encode4(f.headerSize)
	// Copyright
	if f.copyright == "" {
		f.copyright = DefaultWzFileCopyright
	}
	archive.EncodeNTStr(f.copyright)
	// Version
	if f.version == 0 {
		return errors.New("unknown version")
	}
	if f.version != NoHashVersion {
		archive.Encode2(GetWzVersionHashSumTruncated(f.version))
	}
	return nil
}

// WriteMetaNode implements [IWzFile].
func (f *wzFile) WriteMetaNode(ctx context.Context, archive IWzArchive, parnet IWzNode) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Level-order BFS
	dirQueue := make([]IWzNode, 0)
	var nodes []IWzNode
	if parnet == nil {
		nodes = slices.Collect(f.nodeMap.Values())
	} else {
		nodes = parnet.GetChilds()
	}
	nodesLen := len(nodes)
	archive.EncodeVT4(int32(nodesLen))
	for _, node := range nodes {
		node.Serialize(archive)
		nType := node.GetType()
		switch nType { //nolint:exhaustive // not all NodeType cases are needed
		case NodeDirectory:
			dirQueue = append(dirQueue, node)
		case NodeImage, NodeImageRef:
			f.imgQueue = append(f.imgQueue, node)
		default:
			return fmt.Errorf("unknown node type %d at offset %d", nType, archive.GetOffset())
		}
	}
	for _, dir := range dirQueue {
		err := f.WriteMetaNode(ctx, archive, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteDataNode implements [IWzFile].
func (f *wzFile) WriteDataNode(ctx context.Context, archive IWzArchive) error {
	for _, img := range f.imgQueue {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		prop := NewWzProperty(img)
		prop.SetFlag(FlagPropName)
		prop.SetName(string(NodeNameProperty))
		WzDispatchSerialize(prop, archive)
	}
	f.imgQueue = f.imgQueue[:0]
	return nil
}

// #endregion

// Close implements [IWzFile].
func (f *wzFile) Close() {
	if f.stream != nil {
		f.stream.Close()
		f.stream = nil
	}
}
