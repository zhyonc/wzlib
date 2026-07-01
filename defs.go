package wzlib

const (
	DefaultTimeoutSec       int    = 30
	DefaultWzFileIdent      string = "PKG1"
	DefaultWzFileHeaderSize int    = 60
	DefaultWzFileCopyright  string = "Package file v1.0 Copyright 2002 Wizet, ZMS"
	MinVersion              int    = 1
	MaxVersion              int    = 2000
	NoHashVersion           int16  = 777
	NodePathSeparator       string = "/"
)

//nolint:gochecknoglobals // fixed constant
var (
	ASCIIWhitelist = map[rune]struct{}{
		0xB7: {}, // '·'  GMS95  "String.wz/ToolTipHelp.img/Mapobject/200000000/Title" "Weapon·Armor Shop"
		0xE9: {}, // 'é'  GMS95  "Quest.wz/QuestInfo.img/28376/1" "a secret entrée"
		0xFC: {}, // 'ü'  GMS232 "Etc.wz/OXQuiz.img/9/161/d" "Fürnemmen"
		0xA0: {}, // NBSP GMS232 "Quest.wz/QuestInfo.img/16398/name" "[Nova] Festival Nova\u00a0Box Drop Rate"
		0xAD: {}, // SHY  GMS232 "String.wz/PetDialog.img/5000256/c3_s3" "You're a grade\u00adA gourd!"
	}
)
