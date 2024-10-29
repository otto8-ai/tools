package transformers

import (
	"context"
	"fmt"
	"log/slog"

	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

const ExtraMetadataName = "extra_metadata"

type ExtraMetadata struct {
	Metadata map[string]any
}

func (e *ExtraMetadata) Transform(_ context.Context, docs []vs.Document) ([]vs.Document, error) {
	for i, doc := range docs {
		metadata := doc.Metadata
		for k, v := range e.Metadata {
			metadata[k] = v
		}
		docs[i].Metadata = metadata
	}
	return docs, nil
}

func (e *ExtraMetadata) Name() string {
	return ExtraMetadataName
}

const MetadataManipulatorName = "metadata"

type MetadataManipulationOperator string

const (
	MetadataManipulationOperatorAdd    MetadataManipulationOperator = "add"
	MetadataManipulationOperatorUpdate MetadataManipulationOperator = "upsert"
	MetadataManipulationOperatorRemove MetadataManipulationOperator = "remove"
)

type MetadataManipulation struct {
	Operator MetadataManipulationOperator `json:"operator,omitempty" mapstructure:"operator"`
	Key      string                       `json:"key,omitempty" mapstructure:"key"`
	Value    any                          `json:"value,omitempty" mapstructure:"value"`
}

type MetadataManipulator struct {
	Manipulations []MetadataManipulation
}

func (m *MetadataManipulator) Name() string {
	return MetadataManipulatorName
}

func (m *MetadataManipulator) Transform(_ context.Context, docs []vs.Document) ([]vs.Document, error) {
	for i, doc := range docs {
		metadata := doc.Metadata
		if metadata == nil {
			metadata = make(map[string]any)
		}
		slog.Debug("metadata manipulator", "docMetadata", metadata, "manipulations", m.Manipulations)
		for _, manipulation := range m.Manipulations {
			switch manipulation.Operator {
			case MetadataManipulationOperatorAdd:
				if _, exists := metadata[manipulation.Key]; exists {
					return nil, fmt.Errorf("metadata key %q already exists in document", manipulation.Key)
				}
				metadata[manipulation.Key] = manipulation.Value
			case MetadataManipulationOperatorUpdate:
				metadata[manipulation.Key] = manipulation.Value
			case MetadataManipulationOperatorRemove:
				delete(metadata, manipulation.Key)
			}
		}
		slog.Debug("metadata manipulator DONE", "docMetadata", metadata)
		docs[i].Metadata = metadata
	}
	return docs, nil
}
