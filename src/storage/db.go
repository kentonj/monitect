package storage

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/kentonj/monitect/src/conf"
)

func Connect(config *conf.Config) *gorm.DB {
	// make the parent directory if necessary
	parentDir := filepath.Base(config.Database.File)
	err := os.MkdirAll(parentDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
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
	return db
}
