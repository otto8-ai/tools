package index

import (
	"github.com/gptscript-ai/knowledge/pkg/config"
	"time"
)

// Dataset refers to a VectorDB data space.
// @Description Dataset refers to a VectorDB data space.
type Dataset struct {
	ID                       string                      `gorm:"primaryKey" json:"id"`
	EmbeddingsProviderConfig *config.ModelProviderConfig `json:"embeddingsProviderConfig,omitempty" gorm:"serializer:json"`
	Files                    []File                      `gorm:"foreignKey:Dataset;references:ID;constraint:OnDelete:CASCADE;"`
	Metadata                 map[string]any              `json:"metadata,omitempty" gorm:"serializer:json"`
}

type File struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	Dataset   string     `gorm:"primaryKey" json:"dataset"` // Foreign key to Dataset
	Documents []Document `gorm:"foreignKey:FileID,Dataset;references:ID,Dataset;constraint:OnDelete:CASCADE;"`
	// File metadata, commonly used for deduplication
	FileMetadata `json:",inline"`
}

type FileMetadata struct {
	Name         string    `json:"name"`
	AbsolutePath string    `json:"absolute_path"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type Document struct {
	ID      string `gorm:"primaryKey" json:"id"`
	Dataset string `gorm:"primaryKey" json:"dataset"` // Foreign key to Dataset, part of composite primary key with FileID
	FileID  string `gorm:"primaryKey" json:"file_id"` // Foreign key to File, part of composite primary key with Dataset
}
