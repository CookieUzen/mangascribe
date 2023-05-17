package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	time "time"
)

const API = "https://api.mangadex.org"

func searchManga(title string) (Manga, error) {
	// Loading in URL
	fullURL := fmt.Sprintf("%s/manga", API)

	// Send the request
	body, err := requestGET(fullURL, map[string]string{
		"title": title,
		"limit": "1",
	})

	// parse the response as a searchMangaStruct
	var outputManga searchMangaStruct
	err = json.Unmarshal(body, &outputManga)
	if err != nil {
		fmt.Println("Failed to parse response:", err)
		return Manga{}, err
	}

	// Process the response into a Manga struct (get the first result)
	var manga Manga

	manga.id = outputManga.Data[0].ID
	manga.name = outputManga.Data[0].Attributes.Title["en"]

	if err != nil {
		fmt.Println("Failed to get manga:", err)
		return Manga{}, err
	}

	return manga, nil
}

func getChapters(manga *Manga) error {
	// Count the returned chapters versus total chapters
	total := math.MaxInt
	fullURL := fmt.Sprintf("%s/manga/%s/feed", API, manga.id)

	for count := 0; ; {
		// Send the request
		body, err := requestGET(fullURL, map[string]string{
			"offset":               strconv.Itoa(count),
			"translatedLanguage[]": `en`,
		})
		if err != nil {
			return err
		}

		// Parse the response
		var outputManga mangaResponse
		err = json.Unmarshal(body, &outputManga)
		if err != nil {
			fmt.Println("Failed to parse response:", err)
			return err
		}

		// If mangadex is not happy with the request
		if outputManga.Result == "error" {
			return errors.New(outputManga.Response)
		}

		// Init the chapters array only once
		if count == 0 {
			manga.chapters = make([]Chapter, outputManga.Total)
		}

		// Add the chapters to the manga
		for i, chapter := range outputManga.Data {
			chapter := Chapter{
				ID:                 chapter.ID,
				Title:              chapter.Attributes.Title,
				Chapter:            chapter.Attributes.Chapter,
				Volume:             chapter.Attributes.Volume,
				TranslatedLanguage: chapter.Attributes.TranslatedLanguage,
				Pages:              chapter.Attributes.Pages,
			}

			manga.chapters[i+count] = chapter
		}

		// Update the count
		count += len(outputManga.Data)

		// Update the total
		total = outputManga.Total

		// If we have all the chapters, break
		if count >= total {
			break
		}

		// sleep for a millisecond
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

func requestGET(fullURL string, args map[string]string) ([]byte, error) {
	for i := 1; i < 5; i++ {
		client := http.Client{}

		// Loading in URL
		u, err := url.Parse(fullURL)
		if err != nil {
			fmt.Println("Failed to parse URL", err)
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
			fmt.Println("Failed to create request:", err)
			return []byte(""), err
		}

		// Response
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("Failed to send request:", err)
			fmt.Println("Retrying after " + strconv.Itoa(i) + " seconds")
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}

		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return []byte(""), err
		}

		return body, nil
	}

	return []byte(""), errors.New("Failed to send request")
}

func main() {
	// Get the manga
	manga, _ := searchManga("Komi-san")
	getChapters(&manga)

	fmt.Println(manga)
}
