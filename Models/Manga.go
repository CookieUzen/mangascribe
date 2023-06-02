package Models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CookieUzen/mangascribe/Config"
	"github.com/CookieUzen/mangascribe/MangaDex"
	"github.com/CookieUzen/mangascribe/Tools"
	"github.com/golang/glog"
	"gorm.io/gorm"
	"math"
	"strconv"
	"time"
)

type Manga struct {
	gorm.Model
	ID       string `gorm:"primaryKey:false"`
	Name     string
	Chapters []Chapter `gorm:"foreignKey:MangaID"`
	Volumes  []Volume  `gorm:"foreignKey:MangaID"`
}

// Gets a list of all the available chapters for a given Manga struct
func (manga *Manga) getChapters() error {
	// Count the returned Chapters versus total Chapters
	total := math.MaxInt
	fullURL := fmt.Sprintf("%s/manga/%s/feed", Config.API, manga.ID)

	count := 0
	page := 0

	for {
		// Send the request
		glog.Info("Fetching page ", page)
		body, err := Tools.RequestGET(fullURL, map[string]string{
			"offset":               strconv.Itoa(count),
			"translatedLanguage[]": `en`, // TODO: add other options
		})

		// If the request failed, pass the error up
		if err != nil {
			return err
		}

		// Parse the response
		var outputManga MangaDex.MangaResponse
		err = json.Unmarshal(body, &outputManga)
		if err != nil {
			glog.Error("Failed to parse response:", err)
			return err
		}

		// If mangadex is not happy with the request
		if outputManga.Result == "error" {
			err := errors.New(outputManga.Response)
			glog.Error("Mangadex returned an error when fetching chapters: ", err)
			return err
		}

		// Init the Chapters array only once
		if count == 0 {
			manga.Chapters = make([]Chapter, outputManga.Total)
		}

		// Add the Chapters to the manga
		for i, chapter := range outputManga.Data {
			outputChapter := Chapter{
				ID:                 chapter.ID,
				Title:              chapter.Attributes.Title,
				Chapter:            chapter.Attributes.Chapter,
				Volume:             chapter.Attributes.Volume,
				TranslatedLanguage: chapter.Attributes.TranslatedLanguage,
				PageNumber:         chapter.Attributes.Pages,
				// Manga:              manga,
				Pages: make([]Page, chapter.Attributes.Pages),
			}

			// If the chapter name is a float, rename to "Chapter #"
			if _, err := strconv.ParseFloat(outputChapter.Chapter, 64); err == nil {
				outputChapter.Chapter = "Chapter " + outputChapter.Chapter
			}

			// If the volume is empty, set it to "EMPTY_VOLUME_NAME"
			if outputChapter.Volume == "" {
				outputChapter.Volume = Config.EMPTY_VOLUME_NAME
			}

			// If the volume is a float, convert it to "Volume #"
			if _, err := strconv.ParseFloat(outputChapter.Volume, 64); err == nil {
				outputChapter.Volume = "Volume " + outputChapter.Volume
			}

			// Find the Scanlation Group
			for _, relationship := range chapter.Relationships {
				if relationship.Type == "scanlation_group" {
					chapter.Relationships[0] = relationship
					break
				}
			}

			manga.Chapters[i+count] = outputChapter
		}

		// Update the count
		count += len(outputManga.Data)

		// Update the total
		total = outputManga.Total

		// Increment the page
		page++

		// If we have all the Chapters, break
		if count >= total {
			glog.Info("Found ", count, " chapters, Done.")
			break
		}

		// sleep for a millisecond
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

// Sorts the Chapters into Volumes
// Can update the volume after fetching new Chapters
// TODO: prioritize same scanlation group
// TODO: move language selection here
// TODO: skip official chapters with non mangadex links
func (manga *Manga) chapterToVolume() error {
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
func (manga *Manga) download() error {
	// loop through all the volumes
	for _, volume := range manga.Volumes {
		err := volume.download()
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
