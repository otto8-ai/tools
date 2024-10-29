package types

import (
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	golcschema "github.com/hupe1980/golc/schema"
	lcgoschema "github.com/tmc/langchaingo/schema"
)

func FromGolcDocs(docs []golcschema.Document) []vs.Document {
	vsdocs := make([]vs.Document, len(docs))
	for i, doc := range docs {
		vsdocs[i] = vs.Document{
			Content:  doc.PageContent,
			Metadata: doc.Metadata,
		}
	}
	return vsdocs
}

func ToGolcDocs(docs []vs.Document) []golcschema.Document {
	golcdocs := make([]golcschema.Document, len(docs))
	for i, doc := range docs {
		golcdocs[i] = golcschema.Document{
			PageContent: doc.Content,
			Metadata:    doc.Metadata,
		}
	}
	return golcdocs
}

func FromLangchainDocs(docs []lcgoschema.Document) []vs.Document {
	vsdocs := make([]vs.Document, len(docs))
	for i, doc := range docs {
		vsdocs[i] = vs.Document{
			Content:  doc.PageContent,
			Metadata: doc.Metadata,
		}
	}
	return vsdocs
}

func ToLangchainDocs(docs []vs.Document) []lcgoschema.Document {
	lcdocs := make([]lcgoschema.Document, len(docs))
	for i, doc := range docs {
		lcdocs[i] = lcgoschema.Document{
			PageContent: doc.Content,
			Metadata:    doc.Metadata,
		}
	}
	return lcdocs
}
