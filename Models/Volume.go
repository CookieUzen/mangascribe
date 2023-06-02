package Models

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

type Volume struct {
	gorm.Model
	Name     string
	Chapters []Chapter `gorm:"foreignKey:VolumeID"`
	MangaID  uint
}

// This downloads all the chapters in a volume
func (volume *Volume) download() error {
	var volumeName string

	// loops through the chapters
	count := 0
	for _, chapter := range volume.Chapters {
		// Get the volume name
		if count == 0 {
			volumeName = chapter.Volume
		}

		err := chapter.download(false)
		if err != nil {
			errText := fmt.Sprintf("failed to download chapter: %v", err)
			err = errors.New(errText)
			glog.Error(err)
			return err
		}

		count++
	}

	glog.Info("Successfully downloaded volume: ", volumeName)
	return nil
}
