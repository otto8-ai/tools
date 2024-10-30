package index

import (
	"fmt"
	"slices"
)

// SetMetadataField sets a metadata field in the dataset. If the metadata does not exist, it will be created.
// If the metadata field already exists, it will be overwritten.
func (d *Dataset) SetMetadataField(key string, value interface{}) {
	if d.Metadata == nil {
		d.Metadata = make(map[string]interface{})
	}
	d.Metadata[key] = value
	d.cleanMetadata()
}

// ReplaceMetadata replaces the metadata of the dataset with the given metadata.
func (d *Dataset) ReplaceMetadata(metadata map[string]interface{}) {
	d.Metadata = metadata
	d.cleanMetadata()
}

// UpdateMetadata updates the metadata of the dataset with the given metadata.
// If a metadata field already exists, it will be overwritten. If a metadata field does not exist, it will be created.
// Existing metadata fields that are not present in the given metadata will remain unchanged.
func (d *Dataset) UpdateMetadata(metadata map[string]interface{}) {
	for k, v := range metadata {
		d.SetMetadataField(k, v)
	}
	d.cleanMetadata()
}

func (d *Dataset) cleanMetadata() {
	for k, v := range d.Metadata {
		if v == nil || slices.Contains([]string{"", "-", "null", "nil"}, fmt.Sprintf("%v", v)) {
			delete(d.Metadata, k)
		}
	}
}
