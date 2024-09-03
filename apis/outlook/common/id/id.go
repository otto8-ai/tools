package id

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CacheObject struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	OutlookID string `gorm:"index"`
}

var dbPath = path.Join(os.Getenv("GPTSCRIPT_WORKSPACE_DIR"), "outlookcache.db")

func loadDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{LogLevel: logger.Error, IgnoreRecordNotFoundError: true}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load the Outlook cache database: %w", err)
	}

	if err = db.AutoMigrate(&CacheObject{}); err != nil {
		return nil, fmt.Errorf("failed to migrate the Outlook cache database: %w", err)
	}

	return db, nil
}

func GetOutlookID(id string) (string, error) {
	idNum, err := strconv.Atoi(id)
	if err != nil {
		// If the ID does not convert to a number, it's most likely already an Outlook ID, so we just return it back.
		return id, nil
	}

	db, err := loadDB()
	if err != nil {
		return "", err
	}

	var cache CacheObject
	if err = db.First(&cache, idNum).Error; err != nil {
		return "", fmt.Errorf("failed to get the Outlook ID: %w", err)
	}

	return cache.OutlookID, nil
}

func SetOutlookID(outlookID string) (string, error) {
	db, err := loadDB()
	if err != nil {
		return "", err
	}

	// First we try looking for an existing one.
	var existing CacheObject
	if err = db.Where("outlook_id = ?", outlookID).First(&existing).Error; err == nil {
		return strconv.Itoa(int(existing.ID)), nil
	}

	// If it doesn't exist, we create a new one.
	cache := CacheObject{OutlookID: outlookID}
	if err = db.Create(&cache).Error; err != nil {
		return "", fmt.Errorf("failed to set the Outlook ID: %w", err)
	}

	return strconv.Itoa(int(cache.ID)), nil
}
