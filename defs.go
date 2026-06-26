package wzlib

const (
	DefaultTimeoutSec        int    = 30
	DefaultWzFileIdent       string = "PKG1"
	DefaultWzFileHeaderSize  int    = 60
	DefaultWzFileCopyright   string = "Package file v1.0 Copyright 2002 Wizet, ZMS"
	DefaultWzFileMinVersion  int    = 1
	DefaultWzFileMaxVersion  int    = 2000
	DefaultNodePathSeparator string = "/"
)

//nolint:gochecknoglobals // default fixed constant
var (
	WzFilenames = []string{
		"Base",
		"Character",
		"Effect",
		"Etc",
		"Item",
		"Map",
		"Mob",
		"Morph",
		"Npc",
		"Quest",
		"Reactor",
		"Skill",
		"Sound",
		"String",
		"TamingMob",
		"UI",
	}
	ASCIIWhitelist = map[rune]struct{}{
		'·': {},
		'é': {},
	}
)
