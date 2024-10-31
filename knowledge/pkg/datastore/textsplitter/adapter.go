package textsplitter

import (
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	golcschema "github.com/hupe1980/golc/schema"
	lcgosplitter "github.com/tmc/langchaingo/textsplitter"
)

type golcSplitterAdapter struct {
	golcschema.TextSplitter
	name string
}

func (a *golcSplitterAdapter) SplitDocuments(docs []vs.Document) ([]vs.Document, error) {
	golcdocs, err := a.TextSplitter.SplitDocuments(types.ToGolcDocs(docs))
	return types.FromGolcDocs(golcdocs), err
}

func (a *golcSplitterAdapter) Name() string {
	return a.name
}

func FromGolc(splitter golcschema.TextSplitter, name string) types.TextSplitter {
	return &golcSplitterAdapter{splitter, name}
}

func AsGolc(splitter types.TextSplitter) golcschema.TextSplitter {
	return splitter.(*golcSplitterAdapter).TextSplitter
}

// --- langchaingo ---

type langchainSplitterAdapter struct {
	lc   lcgosplitter.TextSplitter
	name string
}

func (a *langchainSplitterAdapter) SplitDocuments(docs []vs.Document) ([]vs.Document, error) {
	lcdocs, err := lcgosplitter.SplitDocuments(a.lc, types.ToLangchainDocs(docs))
	return types.FromLangchainDocs(lcdocs), err
}

func (a *langchainSplitterAdapter) Name() string {
	return a.name
}

func FromLangchain(splitter lcgosplitter.TextSplitter, name string) types.TextSplitter {
	return &langchainSplitterAdapter{splitter, name}
}

func AsLangchain(splitter types.TextSplitter) lcgosplitter.TextSplitter {
	return splitter.(*langchainSplitterAdapter).lc
}
