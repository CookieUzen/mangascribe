package Models

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

type Manga struct {
	gorm.Model
	ID          string `gorm:"primaryKey:false"`
	Name        string
	Chapters    []Chapter `gorm:"foreignKey:MangaID"`
	Volumes     []Volume  `gorm:"foreignKey:MangaID"`
	APIProvider string
}

// Gets a list of all the available chapters for a given Manga struct
func (manga *Manga) GetChapters(API APIProvider, replace bool) error {
	chapters, err := API.FetchChapters(manga.ID)
	if err != nil {
		glog.Error(errors.New("failed to fetch chapters"))
		return err
	}

	if replace {
		manga.Chapters = chapters
		return nil
	}

	// Create a map of the chapters for dedup
	chapterMap := make(map[string]Chapter)
	for _, chapter := range manga.Chapters {
		chapterMap[chapter.ID] = chapter
	}

	// Loop through the new chapters to insert
	for _, chapter := range chapters {
		// Check if the chapter already exists
		_, exists := chapterMap[chapter.ID]

		// If the chapter doesn't exist, add it
		if !exists {
			manga.Chapters = append(manga.Chapters, chapter)
		}
	}

	return nil
}

// Sorts the Chapters into Volumes
// Can update the volume after fetching new Chapters
// TODO: prioritize same scanlation group
// TODO: move language selection here
// TODO: skip official chapters with non mangadex links
func (manga *Manga) ChapterToVolume() error {
	// Init the Volumes array
	manga.Volumes = make([]Volume, 0)

	// Loop through the Chapters
	for _, chapter := range manga.Chapters {
		// chapter.VolumeGroup = &manga.Volumes

		volumeName := chapter.Volume

		// Check if the volume exists
		done := false
		for i, volume := range manga.Volumes {
			if volume.Name == volumeName {
				done = true
				skip := false

				// Check if the chapter already exists
				for _, c := range manga.Volumes[i].Chapters {
					if c.Chapter == chapter.Chapter {
						skip = true
						break
					}
				}
				if skip {
					break
				}

				// Add the chapter to the volume
				manga.Volumes[i].Chapters = append(manga.Volumes[i].Chapters, chapter)
				break
			}
		}
		if done {
			continue
		}

		// Create a new volume and add the chapter if it doesn't exist
		volume := Volume{
			Name:     volumeName,
			Chapters: []Chapter{chapter},
		}
		manga.Volumes = append(manga.Volumes, volume)
	}

	return nil
}

// This downloads all the volumes in a chapter
func (manga *Manga) Download(API APIProvider, datasaver bool) error {
	// loop through all the volumes
	for _, volume := range manga.Volumes {
		err := volume.Download(API, datasaver)
		if err != nil {
			errText := fmt.Sprintf("failed to download volume: %v", err)
			err = errors.New(errText)
			glog.Error(err)
			return err
		}
	}

	glog.Info("Successfully downloaded manga: ", manga.Name)
	return nil
}
