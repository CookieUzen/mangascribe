package Models

type APIProvider interface {
	SearchManga(title string) (Manga, error)
	FetchChapters(id string) ([]Chapter, error)
	FetchChapterDownload(id string, datasaver bool) (string, []string, error)
	GetProvider() string
}
