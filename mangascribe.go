package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

const API = "https://api.mangadex.org"
const EMPTY_VOLUME_NAME = "Extras"

// TODO finetune -v levels

// Creates Manga struct from searching for a title in mangadex
// TODO: Add support for searching for author
// TODO: Add support for searching for tags
// TODO: Add support for returning multiple results
func searchManga(title string) (Manga, error) {
	// Loading in URL
	fullURL := fmt.Sprintf("%s/manga", API)

	// Send the request
	glog.Info("Searching for manga: ", title)
	body, err := requestGET(fullURL, map[string]string{
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
		return Manga{}, err
	}

	// Process the response into a Manga struct (get the first result)
	var manga Manga

	manga.ID = outputManga.Data[0].ID
	manga.Name = outputManga.Data[0].Attributes.Title["en"]

	return manga, nil
}

// Gets a list of all the available chapters for a given Manga struct
func (manga *Manga) getChapters() error {
	// Count the returned Chapters versus total Chapters
	total := math.MaxInt
	fullURL := fmt.Sprintf("%s/manga/%s/feed", API, manga.ID)

	count := 0
	page := 0

	for {
		// Send the request
		glog.Info("Fetching page ", page)
		body, err := requestGET(fullURL, map[string]string{
			"offset":               strconv.Itoa(count),
			"translatedLanguage[]": `en`, // TODO: add other options
		})

		// If the request failed, pass the error up
		if err != nil {
			return err
		}

		// Parse the response
		var outputManga mangaResponse
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
				Pages:              chapter.Attributes.Pages,
				Manga:              manga,
			}

			// If the chapter name is a float, rename to "Chapter #"
			if _, err := strconv.ParseFloat(outputChapter.Chapter, 64); err == nil {
				outputChapter.Chapter = "Chapter " + outputChapter.Chapter
			}

			// If the volume is empty, set it to "EMPTY_VOLUME_NAME"
			if outputChapter.Volume == "" {
				outputChapter.Volume = EMPTY_VOLUME_NAME
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

// Sends a GET request to the given URL with the given args
// Returns the response body as a byte array
// Tries 4 times before giving up, each attempt is n second apart
func requestGET(fullURL string, args map[string]string) ([]byte, error) {
	glog.Info("Sending GET request to ", fullURL, "\nParams: ", args, "\n")
	for i := 1; i < 5; i++ {
		client := http.Client{}

		// Loading in URL
		u, err := url.Parse(fullURL)
		if err != nil {
			glog.Error("Failed to parse URL", err)
			return []byte(""), err
		}

		// query
		q := u.Query()

		// Iterate through the args
		for key, value := range args {
			q.Set(key, value)
		}

		u.RawQuery = q.Encode()

		// GET
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			glog.Error("Failed to create request:", err)
			return []byte(""), err
		}

		// Response
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			glog.Warning("Non 200 response: ", err, "\nRetrying after "+strconv.Itoa(i)+" seconds")
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}

		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			glog.Error("Failed to read response body from request:", err)
			return []byte(""), err
		}

		glog.Info("Successfully sent request, data received")
		return body, nil
	}

	err := errors.New("Failed to send request after 4 attempts")
	glog.Error(err)
	return []byte(""), err
}

// Sorts the Chapters into Volumes
// Can update the volume after fetching new Chapters
// TODO: prioritize same scanlation group
// TODO: move language selection here
func (manga *Manga) chapterToVolume() error {
	// Init the Volumes map
	manga.Volumes = make(map[string]Volume)

	// Loop through the Chapters
	for _, chapter := range manga.Chapters {

		// Add the chapter to the volume
		volumeName := chapter.Volume

		// check if the volume exists
		if _, ok := manga.Volumes[volumeName]; !ok {
			manga.Volumes[volumeName] = make(map[string]Chapter)
		}

		// Check if the chapter exists inside the volume
		if _, ok := manga.Volumes[volumeName][chapter.Chapter]; ok {
			continue
		}

		// Add the chapter to the volume
		manga.Volumes[volumeName][chapter.Chapter] = chapter
	}

	return nil
}

// Downloads the chapters inside the volume map for a manga
// Note that this ignores the chapters array (no duplicate scanlations or languages)
func (chapter Chapter) download(datasaver bool) error {

	// Read the chapter ID
	chapterID := chapter.ID

	// Get the URL from at-home endpoint
	args := map[string]string{
		// "forcePort443": "true",
	}

	linksUnparsed, err := requestGET(API+"/at-home/server/"+chapterID, args)
	if err != nil {
		errText := fmt.Sprintf("Failed to get chapter links: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}

	// Parse the response
	var links downloadChapterRequest
	err = json.Unmarshal(linksUnparsed, &links)
	if err != nil {
		errText := fmt.Sprintf("Failed to parse chapter links: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}

	// Check for 404
	if links.Result == "error" {
		err = errors.New("Failed to get chapter links: chapter id does not exist")
		glog.Error(err)
		return err
	}

	// Get ready for download
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

	// Create the directory
	err, dir := chapter.chapterFolderCreation()
	if err != nil {
		errText := fmt.Sprintf("Failed to create directory: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}

	// Download the files
	for i, link := range linklist {
		filename := fmt.Sprintf("%04d%s", i+1, filepath.Ext(link))

		err = downloadFile(URL+link, filename, dir)
		if err != nil {
			errText := fmt.Sprintf("failed to download link: %v", err)
			err = errors.New(errText)
			glog.Error(err)
			return err
		}
	}

	glog.Info("Successfully downloaded chapter ", chapter.Chapter)
	return nil
}

// Creates the folder for the chapter
// Returns the path to the folder
func (chapter Chapter) chapterFolderCreation() (error, string) {
	// Create the directory
	dirPath := filepath.Join(".", chapter.Manga.Name, chapter.Volume, chapter.Chapter)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		errText := fmt.Sprintf("Failed to create directory: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err, ""
	}

	return nil, dirPath
}

func downloadFile(url string, filename string, directory string) error {
	// Create the file
	file, err := os.Create(path.Join(directory, filename))
	if err != nil {
		errText := fmt.Sprintf("Failed to create file: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}
	defer file.Close()

	for count := 1; count <= 5; count++ {
		// Send HTTP GET request to the URL
		response, err := http.Get(url)
		if err != nil {
			errText := fmt.Sprintf("Failed to send GET request: %w", err)
			err = errors.New(errText)
			debugText := fmt.Sprintf("\nFilename: %w, url: %w\nretrying in %d seconds", filename, url, count)
			glog.Warning(err, debugText)
			time.Sleep(time.Duration(count) * time.Second)
			continue
		}
		defer response.Body.Close()

		// Copy the response body to the file
		_, err = io.Copy(file, response.Body)
		if err != nil {
			errText := fmt.Sprintf("Failed to save file: %w", err)
			err = errors.New(errText)
			glog.Error(err)
			return err
		}

		return nil
	}

	errText := fmt.Sprintf("Failed to download file: ", filename, " from ", url)
	err = errors.New(errText)
	glog.Error(err)
	return err
}

// This downloads all the chapters in a volume
func (volume *Volume) download() error {
	var volumeName string

	// loops through the map
	count := 0
	for _, chapter := range *volume {
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

func main() {
	// For logging flags
	flag.Parse()

	// Get the manga
	manga, _ := searchManga("The Angel Next Door Spoils me rotten")
	manga.getChapters()
	manga.chapterToVolume()

	// test download one chapter
	// manga.Volumes["Volume 1"]["Chapter 1"].download(false)

	// test download manga
	manga.download()

	// Flush logs
	glog.Flush()
}
