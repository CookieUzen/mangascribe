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
	"io"
	"os"
	"path/filepath"
)

type Chapter struct {
	gorm.Model
	ID                 string `gorm:"primaryKey:false"`
	VolumeID           uint
	MangaID            uint
	Volume             string
	Chapter            string
	Title              string
	TranslatedLanguage string
	PageNumber         int
	ScanlationGroup    string
	DownloadPath       string
	Pages              []Page `gorm:"foreignKey:ChapterID"`
}

// Downloads the chapters inside the volume map for a manga
// Note that this ignores the chapters array (no duplicate scanlations or languages)
func (chapter *Chapter) download(datasaver bool) error {

	// Read the chapter ID
	chapterID := chapter.ID

	// Get the URL from at-home endpoint
	args := map[string]string{
		// "forcePort443": "true",
	}

	linksUnparsed, err := Tools.RequestGET(Config.API+"/at-home/server/"+chapterID, args)
	if err != nil {
		errText := fmt.Sprintf("Failed to get chapter links: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}

	// Parse the response
	var links MangaDex.DownloadChapterRequest
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

	// Create the tmp directory to download into
	// tempDir, err := os.MkdirTemp("", chapter.Manga.Name+chapter.Volume+chapter.Chapter)
	tempDir, err := os.MkdirTemp("", chapter.Volume+chapter.Chapter)
	if err != nil {
		errText := fmt.Sprintf("Failed to create temporary directory: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory when done

	// Create the destination directory
	err, dir := chapter.chapterFolderCreation()
	if err != nil {
		errText := fmt.Sprintf("Failed to create directory: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err
	}

	// Download the files into the tmp directory
	for i, link := range linklist {
		filename := fmt.Sprintf("%04d%s", i+1, filepath.Ext(link))

		// Open the origFile for reading and writing
		filePath := filepath.Join(dir, filename)
		origFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			errText := fmt.Sprintf("Failed to open origFile: %w", err)
			glog.Error(errors.New(errText))
			return err
		}
		defer origFile.Close()

		// Check if the origFile in the directory matches the hash in the chapter
		// If it does, skip the download
		if currentHash := chapter.Pages[i].Hash; currentHash != "" {

			// Hash the origFile
			fileHash, err := Tools.HashFile(origFile)
			if err != nil {
				glog.Error(err)
				return err
			}

			if fileHash == currentHash {
				glog.Info("Skipping page ", i+1, " as it already exists")
				continue
			}
		}
		origFile.Close()
		origFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		hash, downloadedFile, err := Tools.DownloadFile(URL+link, filename, tempDir)
		if err != nil {
			errText := fmt.Sprintf("failed to download link: %v", err)
			err = errors.New(errText)
			glog.Error(err)
			return err
		}

		// Update the page
		page := Page{
			Hash:     hash,
			FileName: filename,
			Page:     i,
		}
		chapter.Pages[i] = page

		// Copy the file to the destination directory
		_, err = io.Copy(origFile, downloadedFile)
		if err != nil {
			errText := fmt.Sprintf("Failed to copy file: %w", err)
			err = errors.New(errText)
			glog.Error(err)
			return err
		}
	}

	// Update the chapter
	chapter.DownloadPath = dir

	glog.Info("Successfully downloaded chapter ", chapter.Chapter, " at ", dir)
	return nil
}

// Creates the folder for the chapter
// Returns the path to the folder
func (chapter Chapter) chapterFolderCreation() (error, string) {
	// Create the directory
	// dirPath := filepath.Join(".", chapter.Manga.Name, chapter.Volume, chapter.Chapter)
	dirPath := filepath.Join(".", chapter.Volume, chapter.Chapter)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		errText := fmt.Sprintf("Failed to create directory: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return err, ""
	}

	return nil, dirPath
}
