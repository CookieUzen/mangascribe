package Models

import (
	"errors"
	"fmt"
	"github.com/CookieUzen/mangascribe/Tools"
	"github.com/golang/glog"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
)

type Chapter struct {
	gorm.Model
	VolumeID           uint
	MangaID            uint
	ChapterID          uint `gorm:"primaryKey:true"`

	ID                 string `gorm:"primaryKey:false"`
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
func (chapter *Chapter) Download(API APIProvider, datasaver bool) error {

	// Get URLs
	URL, linklist, err := API.FetchChapterDownload(chapter.ID, datasaver)
	if err != nil {
		glog.Error("Failed to fetch chapter download links")
		return err
	}

	// Create the tmp directory to download into
	// tempDir, err := os.MkdirTemp("", chapter.Manga.Name+chapter.Volume+chapter.Chapter)
	tempDir, err := os.MkdirTemp("", chapter.Volume+chapter.Chapter)
	if err != nil {
		glog.Error("Failed to create temporary directory")
		return err
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			glog.Error(err)
		}
	}(tempDir) // Clean up the temporary directory when done

	// Create the destination directory
	err, dir := chapter.ChapterFolderCreation()
	if err != nil {
		glog.Error("Failed to create directory")
		return err
	}

	// Download the files into the tmp directory
	for i, link := range linklist {
		filename := fmt.Sprintf("%04d%s", i+1, filepath.Ext(link))

		// Open the origFile for reading and writing
		filePath := filepath.Join(dir, filename)
		origFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			glog.Error("Failed to open current file")
			return err
		}
		defer func(origFile *os.File) {
			err := origFile.Close()
			if err != nil {
				glog.Error(err)
			}
		}(origFile)

		// Check if the origFile in the directory matches the hash in the chapter
		// If it does, skip the download
		if currentHash := chapter.Pages[i].Hash; currentHash != "" {

			// Hash the origFile
			fileHash, err := Tools.HashFile(origFile)
			if err != nil {
				glog.Error("Failed to hash file")
				return err
			}

			if fileHash == currentHash {
				glog.Info("Skipping page ", i+1, " as it already exists")
				continue
			}
		}
		// origFile.Close()
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
			err = fmt.Errorf("Failed to copy file: %w", err)
			glog.Error(err)
			return err
		}
	}

	// Update the chapter
	chapter.DownloadPath = dir

	glog.Info("Successfully downloaded chapter ", chapter.Chapter, " at ", dir)
	return nil
}

// ChapterFolderCreation Creates the folder for the chapter
// Returns the path to the folder
func (chapter Chapter) ChapterFolderCreation() (error, string) {
	// Create the directory
	// dirPath := filepath.Join(".", chapter.Manga.Name, chapter.Volume, chapter.Chapter)
	dirPath := filepath.Join(".", chapter.Volume, chapter.Chapter)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		err = fmt.Errorf("Failed to create directory: %w", err)
		glog.Error(err)
		return err, ""
	}

	return nil, dirPath
}
