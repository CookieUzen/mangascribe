package MangaDex

import (
	"encoding/json"
	"fmt"
	"github.com/CookieUzen/mangascribe/Config"
	"github.com/CookieUzen/mangascribe/Models"
	"github.com/CookieUzen/mangascribe/Tools"
	"github.com/golang/glog"
)

// Creates Manga struct from searching for a title in mangadex
// TODO: Add support for searching for author
// TODO: Add support for searching for tags
// TODO: Add support for returning multiple results
func SearchManga(title string) (Models.Manga, error) {
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

	return manga, nil
}
