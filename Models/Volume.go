package Models

import (
	"fmt"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

type Volume struct {
	gorm.Model
	MangaID  uint
	VolumeID uint `gorm:"primaryKey:true"`

	Name     string
	Chapters []Chapter `gorm:"foreignKey:VolumeID"`
}

// This downloads all the chapters in a volume
func (volume *Volume) Download(API APIProvider, datasaver bool) error {
	var volumeName string

	// loops through the chapters
	count := 0
	for _, chapter := range volume.Chapters {
		// Get the volume name
		if count == 0 {
			volumeName = chapter.Volume
		}

		err := chapter.Download(API, datasaver)
		if err != nil {
			err = fmt.Errorf("failed to download chapter %s", chapter.ID)
			glog.Error(err)
			return err
		}

		count++
	}

	glog.Info("Successfully downloaded volume: ", volumeName)
	return nil
}
