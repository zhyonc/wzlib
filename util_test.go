package wzlib_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/zhyonc/wzlib"
)

const testVersion int16 = 95

func TestGetWzVersionHash(t *testing.T) {
	t.Parallel()
	versionHash := wzlib.GetWzVersionHash(testVersion)
	slog.Info("versionHash", "value", fmt.Sprintf("%04X", versionHash))
}

func TestGetWzVersionHashSum(t *testing.T) {
	t.Parallel()
	versionHashSum := wzlib.GetWzVersionHashSum(testVersion)
	slog.Info("versionHashSum", "value", fmt.Sprintf("%04X", versionHashSum))
}
func TestGetWzVersionHashSumTruncated(t *testing.T) {
	t.Parallel()
	versionHashSumTruncated := wzlib.GetWzVersionHashSumTruncated(testVersion)
	slog.Info("versionHashSumTruncated", "value", fmt.Sprintf("%04X", versionHashSumTruncated))
}

func TestGetPossibleVersions(t *testing.T) {
	t.Parallel()
	versionHashSumTruncated := wzlib.GetWzVersionHashSumTruncated(testVersion)
	possibleVersions := wzlib.GetPossibleVersions(versionHashSumTruncated)
	for _, v := range possibleVersions {
		slog.Info("possibleVersion", "value", v)
	}
}
