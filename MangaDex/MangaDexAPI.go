package MangaDex

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CookieUzen/mangascribe/Config"
	"github.com/CookieUzen/mangascribe/Models"
	"github.com/CookieUzen/mangascribe/Tools"
	"github.com/golang/glog"
	"math"
	"strconv"
	"time"
)

type API struct{}

// Creates Manga struct from searching for a title in mangadex
// TODO: Add support for searching for author
// TODO: Add support for searching for tags
// TODO: Add support for returning multiple results
func (API) SearchManga(title string) (Models.Manga, error) {
	// Loading in URL
	fullURL := fmt.Sprintf("%s/manga", Config.API)

	// Send the request
	glog.Info("Searching for manga: ", title)
	body, err := Tools.RequestGET(fullURL, map[string]string{
		"title": title,
		"limit": "1",
	})
	if err != nil {
		glog.Error("Failed to send manga search request:", err)
	}

	// parse the response as a searchMangaStruct
	var outputManga searchMangaStruct
	err = json.Unmarshal(body, &outputManga)
	if err != nil {
		glog.Error("Failed to parse response:", err)
		return Models.Manga{}, err
	}

	// Process the response into a Manga struct (get the first result)
	var manga Models.Manga

	manga.ID = outputManga.Data[0].ID
	manga.Name = outputManga.Data[0].Attributes.Title["en"]

	manga.APIProvider = "API"

	return manga, nil
}

// fetchChapters fetches all the chapters for a given manga
// returns an array of chapters
func (API) FetchChapters(id string) ([]Models.Chapter, error) {
	// Count the returned Chapters versus total Chapters
	total := math.MaxInt
	fullURL := fmt.Sprintf("%s/manga/%s/feed", Config.API, id)

	count := 0
	page := 0

	var output []Models.Chapter

	for {
		// Send the request
		glog.Info("Fetching page ", page)
		body, err := Tools.RequestGET(fullURL, map[string]string{
			"offset":               strconv.Itoa(count),
			"translatedLanguage[]": `en`, // TODO: add other options
		})

		// If the request failed, pass the error up
		if err != nil {
			glog.Error("Failed to send manga search request:", err)
			return nil, err
		}

		// Parse the response
		var outputManga MangaResponse
		err = json.Unmarshal(body, &outputManga)
		if err != nil {
			glog.Error("Failed to parse response:", err)
			return nil, err
		}

		// If mangadex is not happy with the request
		if outputManga.Result == "error" {
			err := errors.New(outputManga.Response)
			glog.Error("Mangadex returned an error when fetching chapters: ", err)
			return nil, err
		}

		// Init the Chapters array only once because we need to know length
		if count == 0 {
			output = make([]Models.Chapter, outputManga.Total)
		}

		// Add the Chapters to the manga
		for i, chapter := range outputManga.Data {
			outputChapter := Models.Chapter{
				ID:                 chapter.ID,
				Title:              chapter.Attributes.Title,
				Chapter:            chapter.Attributes.Chapter,
				Volume:             chapter.Attributes.Volume,
				TranslatedLanguage: chapter.Attributes.TranslatedLanguage,
				PageNumber:         chapter.Attributes.Pages,
				// Manga:              manga,
				Pages: make([]Models.Page, chapter.Attributes.Pages),
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

			output[i+count] = outputChapter
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

	// This returns as many functions as possible
	return output, nil
}

// FetchChapterDownload fetches the download links for a given chapter
// returns a base link plus an array of download links as string
// An optional bool can be passed to download data saver images
func (API) FetchChapterDownload(id string, datasaver bool) (string, []string, error) {
	// Get the URL from at-home endpoint
	args := map[string]string{
		// "forcePort443": "true",
	}

	linksUnparsed, err := Tools.RequestGET(Config.API+"/at-home/server/"+id, args)
	if err != nil {
		glog.Error("Failed to get chapter links")
		return "", nil, err
	}

	// Parse the response
	var links DownloadChapterRequest
	err = json.Unmarshal(linksUnparsed, &links)
	if err != nil {
		glog.Error("Failed to parse chapter links")
		return "", nil, err
	}

	// Check for 404
	if links.Result == "error" {
		glog.Error("Failed to get chapter links: chapter id does not exist")
		return "", nil, err
	}

	// Create the URL
	URL := links.BaseURL
	var linklist []string

	if datasaver {
		URL += "/data-saver/"
		linklist = links.Chapter.DataSaver
	} else {
		URL += "/data/"
		linklist = links.Chapter.Data
	}

	URL += links.Chapter.Hash + "/"

	return URL, linklist, nil
}

func (API) GetProvider() string {
	return "API"
}
