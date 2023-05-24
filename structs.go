package main

import (
	"time"
)

type Volume map[string]Chapter

type Manga struct {
	ID       string
	Name     string
	Chapters []Chapter
	Volumes  map[string]Volume
}

type Chapter struct {
	ID                 string
	Volume             string
	Chapter            string
	Title              string
	TranslatedLanguage string
	PageNumber         int
	ScanlationGroup    string
	Manga              *Manga
	VolumeGroup        *Volume // Pointer to current volume
	DownloadPath       string
	Pages              []Page // 0 indexed
}

type Page struct {
	Page int
	Hash string
}

type getChapterStruct struct {
	ID         string `json:"ID"`
	Type       string `json:"type"`
	Attributes struct {
		Volume             string    `json:"volume"`
		Chapter            string    `json:"chapter"`
		Title              string    `json:"title"`
		TranslatedLanguage string    `json:"translatedLanguage"`
		ExternalURL        string    `json:"externalUrl"`
		PublishAt          time.Time `json:"publishAt"`
		ReadableAt         time.Time `json:"readableAt"`
		CreatedAt          time.Time `json:"createdAt"`
		UpdatedAt          time.Time `json:"updatedAt"`
		Pages              int       `json:"pages"`
		Version            int       `json:"version"`
	} `json:"attributes"`
	Relationships []struct {
		ID   string `json:"ID"`
		Type string `json:"type"`
	} `json:"relationships"`
}

type mangaResponse struct {
	Result   string             `json:"result"`
	Response string             `json:"response"`
	Data     []getChapterStruct `json:"data"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	Total    int                `json:"total"`
}

type searchMangaStruct struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     []struct {
		ID         string `json:"ID"`
		Type       string `json:"type"`
		Attributes struct {
			Title                          map[string]string   `json:"title"`
			AltTitles                      []map[string]string `json:"altTitles"`
			Description                    map[string]string   `json:"description"`
			IsLocked                       bool                `json:"isLocked"`
			Links                          map[string]string   `json:"links"`
			OriginalLanguage               string              `json:"originalLanguage"`
			LastVolume                     string              `json:"lastVolume"`
			LastChapter                    string              `json:"lastChapter"`
			PublicationDemographic         string              `json:"publicationDemographic"`
			Status                         string              `json:"status"`
			Year                           int                 `json:"year"`
			ContentRating                  string              `json:"contentRating"`
			ChapterNumbersResetOnNewVolume bool                `json:"chapterNumbersResetOnNewVolume"`
			LatestUploadedChapter          string              `json:"latestUploadedChapter"`
			Tags                           []struct {
				ID         string `json:"ID"`
				Type       string `json:"type"`
				Attributes struct {
					Name        map[string]string `json:"Name"`
					Description map[string]string `json:"description"`
					Group       string            `json:"group"`
					Version     int               `json:"version"`
				} `json:"attributes"`
				Relationships []struct {
					ID         string `json:"ID"`
					Type       string `json:"type"`
					Related    string `json:"related"`
					Attributes struct {
					} `json:"attributes"`
				} `json:"relationships"`
			} `json:"tags"`
			State     string `json:"state"`
			Version   int    `json:"version"`
			CreatedAt string `json:"createdAt"`
			UpdatedAt string `json:"updatedAt"`
		} `json:"attributes"`
		Relationships []struct {
			ID         string `json:"ID"`
			Type       string `json:"type"`
			Related    string `json:"related"`
			Attributes struct {
			} `json:"attributes"`
		} `json:"relationships"`
	} `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type downloadChapterRequest struct {
	Result  string `json:"result"`
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash      string   `json:"hash"`
		Data      []string `json:"data"`
		DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}
