package wzlib_test

import (
	"errors"
	"fmt"
	"log/slog"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/zhyonc/wzlib"
)

const (
	wzDataDir    string = "./Data"
	wzTempDir    string = "./Temp"
	wzFilenameEx string = ".wz"
)

//nolint:gochecknoglobals // different cases
var (
	wzFileSetting = &wzlib.Setting{
		LocaleRegion: wzlib.GMS,
		MSRegion:     wzlib.GMS,
		MSVersion:    95,
	}
	wzDirSetting = &wzlib.Setting{
		LocaleRegion:     wzlib.GMS,
		MSRegion:         wzlib.GMSCW,
		MSVersion:        1,
		DisableAESCipher: true,
	}
	lazyLoadSetting = &wzlib.Setting{
		IsLazyLoad:   true,
		LocaleRegion: wzlib.GMS,
		MSRegion:     wzlib.GMS,
		MSVersion:    95,
	}
)

func RunWithTime(t *testing.T, work func() error) {
	wzlib.SetLogger(slog.LevelDebug, true)
	start := time.Now()
	slog.Info("Running...", "time", start)
	errs := work()
	if errs != nil {
		slog.Error(errs.Error())
		t.Fail()
		return
	}
	slog.Info("Done", "time", time.Now(), "elapsed", time.Since(start))
}

func loadWzFile(setting *wzlib.Setting, wzFilePath string) (wzlib.IWzFile, error) {
	wzFile := wzlib.NewWzFile(setting)
	err := wzFile.Load(wzFilePath)
	if err != nil {
		return nil, err
	}
	slog.Info("Read wz file",
		"filePath", wzFilePath,
		"size", wzFile.GetSize(),
		"nodesLen", wzFile.GetNodesLen(),
		"elapsed", wzFile.Elapsed(),
	)
	return wzFile, nil
}

func loadWzFiles(setting *wzlib.Setting, wzDir string) (*sync.Map, error) {
	var wzFileMap sync.Map
	var loadedCount int
	var wg sync.WaitGroup
	for _, filename := range wzlib.WzFilenames {
		filenameEx := filename + wzFilenameEx
		wzFilePath := path.Join(wzDir, filenameEx)
		wg.Go(func() {
			wzFile, err := loadWzFile(setting, wzFilePath)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			wzFileMap.Store(filenameEx, wzFile)
			loadedCount++
		})
	}
	wg.Wait()
	if loadedCount < len(wzlib.WzFilenames) {
		return nil, fmt.Errorf("only loaded %d wz files", loadedCount)
	}
	return &wzFileMap, nil
}

func TestLoadWzFile(t *testing.T) {
	t.Parallel()
	const filename string = "Base"
	RunWithTime(t, func() error {
		filenameEx := filename + wzFilenameEx
		wzFilePath := path.Join(wzDataDir, filenameEx)
		wzFile, err := loadWzFile(wzFileSetting, wzFilePath)
		if err != nil {
			return err
		}
		defer wzFile.Close()
		return nil
	})
}

//nolint:paralleltest // high memory usage
func TestLoadWzFiles(t *testing.T) {
	RunWithTime(t, func() error {
		wzFileMap, err := loadWzFiles(wzFileSetting, wzDataDir)
		if err != nil {
			return err
		}
		var rangeErr error
		wzFileMap.Range(func(key, value any) bool {
			filenameEx, ok := key.(string)
			if !ok {
				rangeErr = errors.New("failed to assert string")
				return false
			}
			var wzFile wzlib.IWzFile
			wzFile, ok = value.(wzlib.IWzFile)
			if !ok {
				rangeErr = fmt.Errorf("failed to assert %s wzlib.IWzFile", filenameEx)
				return false
			}
			wzFile.Close()
			return true
		})
		if rangeErr != nil {
			return rangeErr
		}
		return nil
	})
}
func TestLoadWzDir(t *testing.T) {
	t.Parallel()
	RunWithTime(t, func() error {
		var wg sync.WaitGroup
		for _, wzDir := range wzlib.WzFilenames {
			wzFileDir := path.Join(wzDataDir, wzDir)
			paths, err := wzlib.GetFilePaths(wzFileDir, wzFilenameEx)
			if err != nil {
				continue
			}
			for _, wzFilePath := range paths {
				wg.Go(func() {
					var wzFile wzlib.IWzFile
					wzFile, err = loadWzFile(wzDirSetting, wzFilePath)
					if err != nil {
						slog.Error(err.Error())
						return
					}
					defer wzFile.Close()
				})
			}
		}
		wg.Wait()
		return nil
	})
}

func TestFullPathRead(t *testing.T) {
	t.Parallel()
	RunWithTime(t, func() error {
		wzFileMap, err := loadWzFiles(lazyLoadSetting, wzDataDir)
		if err != nil {
			return err
		}
		const filenameEx string = "Map" + wzFilenameEx
		const fullPath string = "Map/Map0/000010000.img/foothold/0/3/1"
		file, ok := wzFileMap.Load(filenameEx)
		if !ok {
			return errors.New("failed to load wz file")
		}
		wzFile, ok := file.(wzlib.IWzFile)
		if !ok {
			return errors.New("failed to assert wzlib.IWzFile")
		}
		var item wzlib.IWzPropertyItem
		item, err = wzFile.GetPropertyItem(fullPath)
		if err != nil {
			return err
		}
		childs := item.GetChilds()
		for _, child := range childs {
			var temp wzlib.IWzPropertyItem
			temp, ok = child.(wzlib.IWzPropertyItem)
			if !ok {
				continue
			}
			var value int32
			value, err = temp.GetInt32()
			if err != nil {
				return err
			}
			slog.Info("Read property item", "key", child.GetName(), "value", value)
		}
		return nil
	})
}

func TestTieredRead(t *testing.T) {
	t.Parallel()
	RunWithTime(t, func() error {
		wzFileMap, err := loadWzFiles(lazyLoadSetting, wzDataDir)
		if err != nil {
			return err
		}
		const filenameEx string = "Map.wz"
		const dirPath string = "Map/Map0"
		const imgPath string = "000010000.img"
		const itemPath string = "foothold/0/3/1"
		file, ok := wzFileMap.Load(filenameEx)
		if !ok {
			return errors.New("failed to load wz file")
		}
		wzFile, ok := file.(wzlib.IWzFile)
		if !ok {
			return errors.New("failed to assert wzlib.IWzFile")
		}
		var dir wzlib.IWzDirectory
		dir, err = wzFile.GetDirectory(dirPath)
		if err != nil {
			return err
		}
		var img wzlib.IWzImage
		img, err = dir.GetImage(imgPath)
		if err != nil {
			return err
		}
		var item wzlib.IWzPropertyItem
		item, err = img.GetPropertyItem(itemPath)
		if err != nil {
			return err
		}
		childs := item.GetChilds()
		for _, child := range childs {
			var temp wzlib.IWzPropertyItem
			temp, ok = child.(wzlib.IWzPropertyItem)
			if !ok {
				continue
			}
			var value int32
			value, err = temp.GetInt32()
			if err != nil {
				return err
			}
			slog.Info("Read property item", "key", child.GetName(), "value", value)
		}
		return nil
	})
}

func TestCreateWzFile(t *testing.T) {
	t.Parallel()
	RunWithTime(t, func() error {
		const filename string = "Base"
		filenameEx := filename + wzFilenameEx
		wzFilePath := path.Join(wzTempDir, filenameEx)
		wzFile := wzlib.NewWzFile(wzFileSetting)
		err := wzFile.Save(wzFilePath, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func TestCopyWzFile(t *testing.T) {
	t.Parallel()
	RunWithTime(t, func() error {
		const srcFilename string = "Base"
		const dstFilename string = "Base"
		srcFilenameEx := srcFilename + wzFilenameEx
		dstFilenameEx := dstFilename + wzFilenameEx
		srcPath := path.Join(wzDataDir, srcFilenameEx)
		dstPath := path.Join(wzTempDir, dstFilenameEx)
		wzFile, err := loadWzFile(wzFileSetting, srcPath)
		if err != nil {
			return err
		}
		defer wzFile.Close()
		err = wzFile.Save(dstPath, false)
		if err != nil {
			return err
		}
		return nil
	})
}

//nolint:paralleltest // high memory usage
func TestCopyWzFiles(t *testing.T) {
	RunWithTime(t, func() error {
		var wg sync.WaitGroup
		for _, filename := range wzlib.WzFilenames {
			srcFilenameEx := filename + wzFilenameEx
			dstFilenameEx := filename + wzFilenameEx
			srcPath := path.Join(wzDataDir, srcFilenameEx)
			dstPath := path.Join(wzTempDir, dstFilenameEx)
			wg.Go(func() {
				wzFile, err := loadWzFile(wzFileSetting, srcPath)
				if err != nil {
					slog.Error(err.Error())
					return
				}
				defer wzFile.Close()
				err = wzFile.Save(dstPath, false)
				if err != nil {
					slog.Error(err.Error())
					return
				}
			})
		}
		wg.Wait()
		return nil
	})
}
