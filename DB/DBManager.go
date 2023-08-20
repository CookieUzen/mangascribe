package DB

import (
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"github.com/golang/glog"
	"github.com/CookieUzen/mangascribe/Config"
	"github.com/CookieUzen/mangascribe/Models"
)

type DBManager struct {
	DB *gorm.DB
}

func Open() DBManager {
	glog.Info("Initializing DBManager")
	db, err := gorm.Open(sqlite.Open(Config.DB_PATH), &gorm.Config{})
	if err != nil {
		glog.Fatalf("Failed to connect to the database: %v", err)
	}

	dbm := DBManager{DB: db}
	dbm.Migrate()

	return dbm
}

// Migrate the database
// Panic if there is an error migrating the database
func (dbm *DBManager) Migrate() {
	err := dbm.DB.AutoMigrate(
		&Models.Manga{},
		&Models.Volume{},
		&Models.Chapter{},
		&Models.Page{},
		&Models.Account{},
		Models.APIKey{},
	)

	if err != nil {
		glog.Fatalf("Failed to migrate the database: %v", err)
	}
}

// Close the database connection
// Panic if there is an error closing the database connection
func (dbm *DBManager) Close() {
	db, err := dbm.DB.DB()
	if err != nil {
		glog.Fatalf("Failed to get underlying sqliteDB: %v", err)
	}

	err = db.Close()
	if err != nil {
		glog.Fatalf("Failed to close the database: %v", err)
	}
}
