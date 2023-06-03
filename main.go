package main

import (
	"flag"
	"fmt"
	"github.com/CookieUzen/mangascribe/MangaDex"
	"github.com/CookieUzen/mangascribe/Models"
	"github.com/golang/glog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TODO finetune -v levels
// TODO fix glog error passing

func main() {
	// For logging flags
	flag.Parse()

	// Connect to the database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Models.Manga{}, &Models.Volume{}, &Models.Chapter{}, &Models.Page{})

	// Insert a manga into the database
	manga, err := MangaDex.API{}.SearchManga("komi-san")
	if err != nil {
		panic(err)
	}

	manga.GetChapters(MangaDex.API{}, true)

	db.Create(&manga.Chapters)
	db.Create(&manga)

	var manga2 Models.Manga
	db.First(&manga2, "id = a96676e5-8ae2-425e-b549-7f15dd34a6d8")

	fmt.Println(manga2.Name)

	// Flush logs
	glog.Flush()
}
