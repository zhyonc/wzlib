package wzlib_test

import (
	"errors"
	"fmt"
	"image/png"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/zhyonc/wzlib"
)

// #region Common

const (
	kWzDataDir      string = "Data"
	kWzTempDir      string = "Temp"
	kSingleSubDir   string = "Single"
	kSplitSubDir    string = "Split"
	kFolderSubDir   string = "Folder"
	kCanvasSubDir   string = "_Canvas"
	kCanvasFilename string = "Canvas"
	kImageFilename  string = "Image"
	kAudioFilename  string = "Audio"
	kFontFilename   string = "Font"
)

const (
	kWzExt  string = ".wz"
	kPNGExt string = ".png"
	kMP3Ext string = ".mp3"
	KWAVExt string = ".wav"
)

type wzFileTestTable struct {
	name    string
	setting *wzlib.Setting
}

//nolint:gochecknoglobals // different cases
var (
	customWzSetting = &wzlib.Setting{
		LocaleRegion: wzlib.GMS,
		MSRegion:     wzlib.GMS,
		MSVersion:    95,
	}
	folderWzSetting = &wzlib.Setting{
		IsLazyLoad:       true,
		LocaleRegion:     wzlib.GMS,
		MSRegion:         wzlib.GMSCW,
		MSVersion:        1,
		DisableAESCipher: true,
	}
	singleWzSetting = &wzlib.Setting{
		IsLazyLoad:   true,
		LocaleRegion: wzlib.GMS,
		MSRegion:     wzlib.GMS,
		MSVersion:    95,
	}
	splitWzSetting = &wzlib.Setting{
		IsLazyLoad:       true,
		LocaleRegion:     wzlib.GMS,
		MSRegion:         wzlib.GMS,
		MSVersion:        777,
		DisableAESCipher: true,
	}
	wzFileTestTables = []wzFileTestTable{
		{
			name:    kFolderSubDir,
			setting: folderWzSetting,
		},
		{
			name:    kSingleSubDir,
			setting: singleWzSetting,
		},
		{
			name:    kSplitSubDir,
			setting: splitWzSetting,
		},
	}
	canvasDataFullPaths = []string{
		"ARGB.img/ARGB8888",
		"ARGB.img/ARGB4444",
		"ARGB.img/ARGB1555",
		"ARGB.img/RGB565",
		"DXT.img/DXT3",
		"DXT.img/DXT5",
	}
	soundDataFullPaths = []string{
		"BGM.img/MP3",
		"BGM.img/PCM",
		"BGM.img/MP3Ex",
	}
)

// #region Helper

func loadWzFile(setting *wzlib.Setting, wzFilePath string) (wzlib.IWzFile, error) {
	wzFile := wzlib.NewWzFile(setting)
	err := wzFile.Load(wzFilePath)
	if err != nil {
		return nil, err
	}
	slog.Info("Read wz file",
		"filePath", wzFilePath,
		"size", wzFile.GetSize(),
		"nodesLen", wzFile.GetMetaNodeLen(),
		"elapsed", wzFile.Elapsed(),
	)
	return wzFile, nil
}

func loadWzDir(setting *wzlib.Setting, dataDir string) (*sync.Map, error) {
	var wzFileMap sync.Map
	var wg sync.WaitGroup
	var pathCount int32
	var loadedCount int32
	paths, err := wzlib.GetFilePaths(dataDir, kWzExt)
	if err != nil {
		return nil, err
	}
	pathCount = int32(len(paths))
	for _, path := range paths {
		wg.Go(func() {
			wzFile, err := loadWzFile(setting, path)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			filenameExt := filepath.Base(path)
			if strings.Contains(filenameExt, kCanvasFilename) {
				dirName := filepath.Base(filepath.Dir(filepath.Dir(path)))
				filenameExt = fmt.Sprintf("%s/%s", dirName, filenameExt)
			}
			wzFileMap.Store(filenameExt, wzFile)
			atomic.AddInt32(&loadedCount, 1)
		})
	}
	wg.Wait()
	if loadedCount < pathCount {
		return nil, fmt.Errorf("only loaded %d/%d wz files", loadedCount, pathCount)
	}
	return &wzFileMap, nil
}

func getWzFile(wzFileMap *sync.Map, filenameExt string) (wzlib.IWzFile, error) {
	file, ok := wzFileMap.Load(filenameExt)
	if !ok {
		return nil, fmt.Errorf("failed to load wz file %s", filenameExt)
	}
	wzFile, ok := file.(wzlib.IWzFile)
	if !ok {
		return nil, errors.New("failed to assert wzlib.IWzFile")
	}
	return wzFile, nil
}

func extractImage(setting *wzlib.Setting, wzFilePath string, fullPath string, outPath string) error {
	var wzFile wzlib.IWzFile
	wzFile, err := loadWzFile(setting, wzFilePath)
	if err != nil {
		return err
	}
	node, err := wzFile.GetNode(fullPath)
	if err != nil {
		return err
	}
	item, err := node.ParseItem()
	if err != nil {
		return err
	}
	canvas, err := item.GetCanvas()
	if err != nil {
		return err
	}
	rgbaImg, err := canvas.ExtractImage()
	if err != nil {
		return err
	}
	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, rgbaImg)
}

func readSound(setting *wzlib.Setting, wzFilePath string, fullPath string) (wzlib.IWzSound, error) {
	var wzFile wzlib.IWzFile
	wzFile, err := loadWzFile(setting, wzFilePath)
	if err != nil {
		return nil, err
	}
	node, err := wzFile.GetNode(fullPath)
	if err != nil {
		return nil, err
	}
	item, err := node.ParseItem()
	if err != nil {
		return nil, err
	}
	sound, err := item.GetSound()
	if err != nil {
		return nil, err
	}
	return sound, nil
}

func extractAudio(setting *wzlib.Setting, wzFilePath string, fullPath string, outPath string) error {
	sound, err := readSound(setting, wzFilePath, fullPath)
	if err != nil {
		return err
	}
	soundData, err := sound.ExtractAudio()
	if err != nil {
		return err
	}
	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(soundData)
	return err
}

func extractFont(setting *wzlib.Setting, wzFilePath string, fullPath string, outPath string) error {
	sound, err := readSound(setting, wzFilePath, fullPath)
	if err != nil {
		return err
	}
	blob, err := sound.ExtractBlob()
	if err != nil {
		return err
	}
	file, err := os.Create(outPath + wzlib.GetFontExt(blob))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(blob)
	return err
}

func extractSkel(setting *wzlib.Setting, wzFilePath string, fullPath string, outPath string) error {
	sound, err := readSound(setting, wzFilePath, fullPath)
	if err != nil {
		return err
	}
	blob, err := sound.ExtractBlob()
	if err != nil {
		return err
	}
	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(blob)
	return err
}

// #endregion

// #region File Cases

func TestLoadCustomWzFile(t *testing.T) {
	t.Parallel()
	const filename string = "Base"
	const filenameExt = filename + kWzExt
	wzFilePath := path.Join(kWzDataDir, kSplitSubDir, filenameExt)
	wzFile, err := loadWzFile(customWzSetting, wzFilePath)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer wzFile.Close()
}

func TestCopyCustomWzFile(t *testing.T) {
	t.Parallel()
	const filename string = "Base"
	const filenameExt = filename + kWzExt
	srcPath := path.Join(kWzDataDir, kSingleSubDir, filenameExt)
	dstPath := path.Join(kWzTempDir, kSingleSubDir, filenameExt)
	wzFile, err := loadWzFile(customWzSetting, srcPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer wzFile.Close()
	err = wzFile.Save(dstPath, false)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestLoadWzFile(t *testing.T) {
	t.Parallel()
	const filename string = "Base"
	const canvasName string = ""
	var filenameExt string
	if canvasName == "" {
		filenameExt = filename + kWzExt
	} else {
		filenameExt = canvasName + kWzExt
	}
	for _, tt := range wzFileTestTables {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var wzFilePath string
			if tt.name == kFolderSubDir {
				if canvasName == "" {
					wzFilePath = path.Join(kWzDataDir, tt.name, filename, filenameExt)
				} else {
					wzFilePath = path.Join(kWzDataDir, tt.name, filename, kCanvasSubDir, filenameExt)
				}
			} else {
				wzFilePath = path.Join(kWzDataDir, tt.name, filenameExt)
			}
			wzFile, err := loadWzFile(tt.setting, wzFilePath)
			if err != nil {
				t.Fatal(err.Error())
			}
			defer wzFile.Close()
		})
	}
}

func TestLoadWzDir(t *testing.T) {
	t.Parallel()
	for _, tt := range wzFileTestTables {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wzDataDir := path.Join(kWzDataDir, tt.name)
			wzFileMap, err := loadWzDir(tt.setting, wzDataDir)
			if err != nil {
				t.Fatal(err.Error())
			}
			var rangeErr error
			wzFileMap.Range(func(key, value any) bool {
				filenameExt, ok := key.(string)
				if !ok {
					rangeErr = errors.New("failed to assert string")
					return false
				}
				var wzFile wzlib.IWzFile
				wzFile, ok = value.(wzlib.IWzFile)
				if !ok {
					rangeErr = fmt.Errorf("failed to assert %s wzlib.IWzFile", filenameExt)
					return false
				}
				defer wzFile.Close()
				metaNodeKeys := wzFile.GetMetaNodeKeys()
				t.Logf("Range key %s and metaNodeKeysLen %d", filenameExt, len(metaNodeKeys))
				return true
			})
			if rangeErr != nil {
				t.Fatal(rangeErr.Error())
			}
		})
	}
}

func TestCreateWzFile(t *testing.T) {
	t.Parallel()
	const filename string = "Base"
	const canvasName string = ""
	var filenameExt string
	if canvasName == "" {
		filenameExt = filename + kWzExt
	} else {
		filenameExt = canvasName + kWzExt
	}
	for _, tt := range wzFileTestTables {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var wzFilePath string
			if tt.name == kFolderSubDir {
				if canvasName == "" {
					wzFilePath = path.Join(kWzTempDir, tt.name, filename, filenameExt)
				} else {
					wzFilePath = path.Join(kWzTempDir, tt.name, filename, kCanvasSubDir, filenameExt)
				}
			} else {
				wzFilePath = path.Join(kWzTempDir, tt.name, filenameExt)
			}
			wzFile := wzlib.NewWzFile(tt.setting)
			err := wzFile.Save(wzFilePath, true)
			if err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

// #endregion

// #region Item Cases

func TestFullPathRead(t *testing.T) {
	t.Parallel()
	filenameExtMap := map[string]string{
		kFolderSubDir: "Map0_000" + kWzExt,
		kSingleSubDir: "Map" + kWzExt,
		kSplitSubDir:  "Map002" + kWzExt,
	}
	fullPathMap := map[string]string{
		kFolderSubDir: "000000001.img/foothold/0/3/1",
		kSingleSubDir: "Map/Map0/000010000.img/foothold/0/3/1",
		kSplitSubDir:  "Map/Map0/000010000.img/foothold/0/2/1",
	}
	for _, tt := range wzFileTestTables {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wzDataDir := path.Join(kWzDataDir, tt.name)
			wzFileMap, err := loadWzDir(tt.setting, wzDataDir)
			if err != nil {
				t.Fatal(err.Error())
			}
			wzFile, err := getWzFile(wzFileMap, filenameExtMap[tt.name])
			if err != nil {
				t.Fatal(err.Error())
			}
			item, err := wzFile.GetPropertyItem(fullPathMap[tt.name])
			if err != nil {
				t.Fatal(err.Error())
			}
			childs := item.GetChilds()
			for _, child := range childs {
				subItem, err := child.ParseItem()
				if err != nil {
					t.Error(err.Error())
					continue
				}
				value, err := subItem.GetInt32()
				if err != nil {
					t.Fatal(err.Error())
				}
				t.Logf("Read property item key %s and value %v", child.GetName(), value)
			}
		})
	}
}

//nolint:gocognit // reduce cognitive complexity
func TestTieredRead(t *testing.T) {
	t.Parallel()
	filenameExtMap := map[string]string{
		kFolderSubDir: "Map0_000" + kWzExt,
		kSingleSubDir: "Map" + kWzExt,
		kSplitSubDir:  "Map002" + kWzExt,
	}
	dirPathMap := map[string]string{
		kFolderSubDir: "",
		kSingleSubDir: "Map/Map0",
		kSplitSubDir:  "Map/Map0",
	}
	imgPathMap := map[string]string{
		kFolderSubDir: "000000001.img",
		kSingleSubDir: "000010000.img",
		kSplitSubDir:  "000010000.img",
	}
	itemPathMap := map[string]string{
		kFolderSubDir: "foothold/0/3/1",
		kSingleSubDir: "foothold/0/3/1",
		kSplitSubDir:  "foothold/0/2/1",
	}
	for _, tt := range wzFileTestTables {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wzDataDir := path.Join(kWzDataDir, tt.name)
			wzFileMap, err := loadWzDir(tt.setting, wzDataDir)
			if err != nil {
				t.Fatal(err.Error())
			}
			wzFile, err := getWzFile(wzFileMap, filenameExtMap[tt.name])
			if err != nil {
				t.Fatal(err.Error())
			}
			dirPath, ok := dirPathMap[tt.name]
			if !ok {
				t.Fatal("unknown dir path")
			}
			var img wzlib.IWzImage
			if dirPath != "" {
				dir, err := wzFile.GetDirectory(dirPath)
				if err != nil {
					t.Fatal(err.Error())
				}
				img, err = dir.GetImage(imgPathMap[tt.name])
				if err != nil {
					t.Fatal(err.Error())
				}
			} else {
				img, err = wzFile.GetImage(imgPathMap[tt.name])
				if err != nil {
					t.Fatal(err.Error())
				}
			}
			item, err := img.GetPropertyItem(itemPathMap[tt.name])
			if err != nil {
				t.Fatal(err.Error())
			}
			childs := item.GetChilds()
			for index, child := range childs {
				subItem, err := child.ParseItem()
				if err != nil {
					slog.Error(err.Error(), "index", index)
					continue
				}
				value, err := subItem.GetInt32()
				if err != nil {
					t.Fatal(err.Error())
				}
				slog.Info("Read property item", "key", child.GetName(), "value", value)
			}
		})
	}
}

// #region Canvas Cases

func TestExtractImage(t *testing.T) {
	t.Parallel()
	const filenameExt string = kImageFilename + kWzExt
	wzFilePath := path.Join(kWzDataDir, filenameExt)
	for _, fullPath := range canvasDataFullPaths {
		outPath := path.Join(kWzTempDir, filepath.Base(fullPath)+kPNGExt)
		err := extractImage(singleWzSetting, wzFilePath, fullPath, outPath)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
	}
}

// #endregion

// #region Sound Cases
func TestExtractAudio(t *testing.T) {
	t.Parallel()
	const filenameExt string = kAudioFilename + kWzExt
	wzFilePath := path.Join(kWzDataDir, filenameExt)
	for _, fullPath := range soundDataFullPaths {
		filename := filepath.Base(fullPath)
		var ext string
		if filename == "MP3" {
			ext = kMP3Ext
		} else {
			ext = KWAVExt
		}
		outPath := path.Join(kWzTempDir, filepath.Base(fullPath)+ext)
		err := extractAudio(singleWzSetting, wzFilePath, fullPath, outPath)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
	}
}

func TestExtractFont(t *testing.T) {
	t.Parallel()
	const filenameExt string = kAudioFilename + kWzExt
	const fullPath string = "YUNGOTHIC250.img/FONT_DATA"
	wzFilePath := path.Join(kWzDataDir, filenameExt)
	outPath := path.Join(kWzTempDir, kFontFilename)
	err := extractFont(singleWzSetting, wzFilePath, fullPath, outPath)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestExtractSkel(t *testing.T) {
	t.Parallel()
	const filenameExt string = kAudioFilename + kWzExt
	const fullPath string = "8881000.img/8881000"
	wzFilePath := path.Join(kWzDataDir, filenameExt)
	outPath := path.Join(kWzTempDir, "8881000.skel")
	err := extractSkel(singleWzSetting, wzFilePath, fullPath, outPath)
	if err != nil {
		t.Fatal(err.Error())
	}
}

// #endregion
