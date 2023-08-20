package Models

import (
	"gorm.io/gorm"
	"time"
	"crypto/rand"
	"encoding/hex"
	"github.com/golang/glog"
	"fmt"
)

type APIKey struct {
	gorm.Model
	ID			uint		`json:"id" gorm:"primaryKey"`
	AccountID	uint		`json:"account_id"`
	Key			string		`json:"key"`
	ExpiresAt	time.Time	`json:"expires_at"`
}

// Generate a new API key
func (account *Account) GenerateAPIKey(duration time.Duration) (*APIKey, error) {
	bytes := make([]byte, 16) // 128-bit key
	if _, err := rand.Read(bytes); err != nil {
		error := fmt.Errorf("Error generating API key: %v", err)
		glog.Error(error)
		return nil, err
	}

	return &APIKey{
		Key:       hex.EncodeToString(bytes),
		ExpiresAt: time.Now().Add(duration),
	}, nil
}

// Check if an API key is expired
func (key *APIKey) IsExpired() bool {
	return key.ExpiresAt.Before(time.Now())
}
