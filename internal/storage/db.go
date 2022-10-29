package storage

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/kentonj/monitect/internal/conf"
)

// connect, run some pre-initialization if necessary and return a pointer to the database
func Connect(config *conf.Config) *gorm.DB {
	// make the parent directory if necessary
	parentDir := filepath.Dir(config.Database.File)
	err := os.MkdirAll(parentDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	log.Printf("using database file %s", config.Database.File)
	db, err := gorm.Open(sqlite.Open(config.Database.File), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if config.Database.InitSQL != "" {
		log.Println("executing the following initialization sql \n", config.Database.InitSQL)
		res := db.Exec(config.Database.InitSQL)
		if res.Error != nil {
			panic(err)
		}
	}
	if config.Database.Debug {
		db = db.Debug()
	}
	return db
}
