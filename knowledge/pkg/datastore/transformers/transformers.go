package transformers

import (
	"fmt"

	dstypes "github.com/gptscript-ai/knowledge/pkg/datastore/types"
)

var TransformerMap = map[string]dstypes.DocumentTransformer{
	ExtraMetadataName:               &ExtraMetadata{},
	FilterMarkdownDocsNoContentName: &FilterMarkdownDocsNoContent{},
	KeywordExtractorName:            &KeywordExtractor{},
	MetadataManipulatorName:         &MetadataManipulator{},
}

func GetTransformer(name string) (dstypes.DocumentTransformer, error) {
	transformer, ok := TransformerMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown transformer %q", name)
	}
	return transformer, nil
}
