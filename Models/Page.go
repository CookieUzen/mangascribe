package Models

import "gorm.io/gorm"

type Page struct {
	gorm.Model
	ChapterID uint
	Page      int
	FileName  string
	Hash      string
}
