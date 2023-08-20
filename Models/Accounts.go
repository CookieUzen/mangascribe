package Models

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	ID			uint		`json:"id" gorm:"primaryKey"`
	Username	string		`json:"username" gorm:"unique"`
	Password	string		`json:"password"`
	Email		string		`json:"email"`
	API_Keys	[]APIKey	`json:"api_keys" gorm:"foreignKey:AccountID"`
}

type NewAccountRequest struct {
	Username	string		`json:"username" binding:"required"`
	Password	string		`json:"password" binding:"required"`
	Email		string		`json:"email" binding:"required"`
}

type LoginRequest struct {
	Identifier	string		`json:"email"`
	Password	string		`json:"password" binding:"required"`
}
