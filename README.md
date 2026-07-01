# wzlib
wzlib is MapleStory wz file binary parser package implemented in Golang

## Installation
 `$ go get github.com/zhyonc/wzlib@latest`

## Quick Start
```golang
package main

import (
	"log/slog"
	"path"

	"github.com/zhyonc/wzlib"
)

const (
	wzDataDir  string = "./Data"
	wzFilename string = "Base.wz"
)

func main() {
	setting := &wzlib.Setting{
		IsLazyLoad:   true,
		LocaleRegion: wzlib.GMS,
		MSRegion:     wzlib.GMS,
		MSVersion:    95,
	}
	wzFilePath := path.Join(wzDataDir, wzFilename)
	wzFile := wzlib.NewWzFile(setting)
	err := wzFile.Load(wzFilePath)
	if err != nil {
		panic(err)
	}
	defer wzFile.Close()
	slog.Info("Read wz file",
		"filename", wzFilename,
		"size", wzFile.GetSize(),
		"nodesLen", wzFile.GetNodesLen(),
		"elapsed", wzFile.Elapsed(),
	)
}
```

## Layout
### Data
Directory for storing original wz files
- Folder: Files organized in multiple directories (such as `GMSCW1`)
- Single: Each file exists as a single file (such as `GMS95`)
- Split: Same-named files split into multiple parts with index suffixes (such as `GMS232`)
- Audio.wz: Contains `Sound` nodes in various formats for testing (export using `GMS95`)
- Image.wz: Contains `Canvas` nodes in various formats for testing (export using `GMS95`)
- [Testing resources](https://mega.nz/folder/alkF2JqC#A3m3upGLrlzuFxUUTWKCRg)
### Temp
Temporary directory for storing test results
- Folder: Testing result from `Data/Folder`
- Single: Testing result from `Data/Single`
- Split: Testing result from `Data/Split`

## How to use
Please refer to [wzlib_test](https://github.com/zhyonc/wzlib/blob/main/wzlib_test.go) for more usage examples 
- CopyCustomWzFile: Reads an existing wz file and writes its structure and offsets to another wz file
- FullPathRead: Reads a property item directly using its `FullPath`
- TieredRead:  Reads a property item step by step through `DirPath-ImgPath-ItemPath` layers
- ExtractImage: 
	- Reads canvas data including `ARGB8888/ARGB4444/ARGB1555/RGB565/DXT3/DXT5` format
	- Converts the canvas data to `RGBA8888` pixel data
	- Writes the pixel data to `.png` file
- ExtractAudio:
	- Reads sound data including `MP3/PCM-in-WAV/MP3-in-WAV`
	- Writes the sound data to `.mp3/.wav` file
- ExtractFont: Writes the sound data to `.ttf/.otf/.bin` file
- ExtractSkel: Writes the sound data to `.skel` file

## Setting
- IsLazyLoad: Reads `MetaNode` at startup and parses `DataNode` on access to minimize memory usage
- LocaleRegion: Language Regions including `EUCKR(KMS)/ShiftJIS(JMS)/GBK(CMS)/Big5(TMS)/Windows1252(GMS)`
- MSRegion: MapleStory Regions including `GMSCW(1)/KMS(1)`/`KMST(2)`/`JMS(3)`/`CMS(4)`/`CMST(5)`/`TMS(6)`/`MSEA(7)`/`GMS(8)`/`BMS(9)`
- MSVersion: MapleStory Client Version
- DisableAESCipher (optional): Some old clients did not apply `AESCipher` to wz files
- IVKey (optional): A 4-byte array used for `AESCipher`
	- If provided directly, this key is used as‑is
	- If not set (first byte is zero), the key is derived according to the `MSRegion`
- AESKey (optional): A 32-byte array used for `AESCipher`
	- If provided directly, this key is used as‑is
	- If not set (first byte is zero), the `AESKeyDefault` is used

## Refer
- [WzWiki](https://mapleref.fandom.com/wiki/WZ)
- [WzComparerR2](https://github.com/Kagamia/WzComparerR2)
- [Arnuh Hasuite](https://archive.softwareheritage.org/browse/origin/directory/?origin_url=https://github.com/Arnuh/Harepacker-resurrected)