package helpers

import (
	"io/ioutil"
	"talisman/gitrepo"
	mockchecksumcalculator "talisman/internal/mock/checksumcalculator"
	mockutility "talisman/internal/mock/utility"
	"talisman/talismanrc"

	"github.com/golang/mock/gomock"
	logr "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"testing"
)

func init() {
	logr.SetOutput(ioutil.Discard)
}
func TestChecksumCompare_IsScanNotRequired(t *testing.T) {

	t.Run("should return false if talismanrc is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSHA256Hasher := mockutility.NewMockSHA256Hasher(ctrl)
		ignoreConfig := &talismanrc.TalismanRC{
			IgnoreConfigs: []talismanrc.IgnoreConfig{},
		}
		cc := NewChecksumCompare(nil, mockSHA256Hasher, ignoreConfig)
		addition := gitrepo.Addition{Path: "some.txt"}
		mockSHA256Hasher.EXPECT().CollectiveSHA256Hash([]string{string(addition.Path)}).Return("somesha")

		required := cc.IsScanNotRequired(addition)

		assert.False(t, required)
	})

	t.Run("should loop through talismanrc configs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSHA256Hasher := mockutility.NewMockSHA256Hasher(ctrl)
		checksumCalculator := mockchecksumcalculator.NewMockChecksumCalculator(ctrl)
		ignoreConfig := talismanrc.TalismanRC{
			IgnoreConfigs: []talismanrc.IgnoreConfig{
				&talismanrc.FileIgnoreConfig{
					FileName: "some.txt",
					Checksum: "sha1",
				},
			},
		}
		cc := NewChecksumCompare(checksumCalculator, mockSHA256Hasher, &ignoreConfig)
		addition := gitrepo.Addition{Name: "some.txt"}
		mockSHA256Hasher.EXPECT().CollectiveSHA256Hash([]string{string(addition.Path)}).Return("somesha")
		checksumCalculator.EXPECT().CalculateCollectiveChecksumForPattern("some.txt").Return("sha1")

		required := cc.IsScanNotRequired(addition)

		assert.True(t, required)
	})

	t.Run("should find any matching talismanrc config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSHA256Hasher := mockutility.NewMockSHA256Hasher(ctrl)
		checksumCalculator := mockchecksumcalculator.NewMockChecksumCalculator(ctrl)
		ignoreConfig := talismanrc.TalismanRC{
			IgnoreConfigs: []talismanrc.IgnoreConfig{
				&talismanrc.FileIgnoreConfig{
					FileName: "some.txt",
					Checksum: "sha1",
				},
				&talismanrc.FileIgnoreConfig{
					FileName: "some.txt",
					Checksum: "recent-sha1",
				},
			},
		}
		cc := NewChecksumCompare(checksumCalculator, mockSHA256Hasher, &ignoreConfig)
		addition := gitrepo.Addition{Name: "some.txt"}
		mockSHA256Hasher.EXPECT().CollectiveSHA256Hash([]string{string(addition.Path)}).Return("somesha")
		checksumCalculator.EXPECT().CalculateCollectiveChecksumForPattern("some.txt").Return("recent-sha1").Times(2)

		required := cc.IsScanNotRequired(addition)

		assert.True(t, required)
	})

	t.Run("should find checksum talismanrc config only", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSHA256Hasher := mockutility.NewMockSHA256Hasher(ctrl)
		checksumCalculator := mockchecksumcalculator.NewMockChecksumCalculator(ctrl)
		ignoreConfig := talismanrc.TalismanRC{
			IgnoreConfigs: []talismanrc.IgnoreConfig{
				&talismanrc.FileIgnoreConfig{
					FileName:        "*.txt",
					AllowedPatterns: []string{"key"},
				},
				&talismanrc.FileIgnoreConfig{
					FileName: "some.txt",
					Checksum: "sha1",
				},
			},
		}
		cc := NewChecksumCompare(checksumCalculator, mockSHA256Hasher, &ignoreConfig)
		addition := gitrepo.Addition{Name: "some.txt"}
		mockSHA256Hasher.EXPECT().CollectiveSHA256Hash([]string{string(addition.Path)}).Return("somesha")
		checksumCalculator.EXPECT().CalculateCollectiveChecksumForPattern("*.txt").Return("sha1")
		checksumCalculator.EXPECT().CalculateCollectiveChecksumForPattern("some.txt").Return("sha1")

		required := cc.IsScanNotRequired(addition)

		assert.True(t, required)
	})
}
