package textsplitter

import (
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	golcschema "github.com/hupe1980/golc/schema"
	lcgosplitter "github.com/tmc/langchaingo/textsplitter"
)

type golcSplitterAdapter struct {
	golcschema.TextSplitter
}

func (a *golcSplitterAdapter) SplitDocuments(docs []vs.Document) ([]vs.Document, error) {
	golcdocs, err := a.TextSplitter.SplitDocuments(types.ToGolcDocs(docs))
	return types.FromGolcDocs(golcdocs), err
}

func FromGolc(splitter golcschema.TextSplitter) types.TextSplitter {
	return &golcSplitterAdapter{splitter}
}

func AsGolc(splitter types.TextSplitter) golcschema.TextSplitter {
	return splitter.(*golcSplitterAdapter).TextSplitter
}

// --- langchaingo ---

type langchainSplitterAdapter struct {
	lc lcgosplitter.TextSplitter
}

func (a *langchainSplitterAdapter) SplitDocuments(docs []vs.Document) ([]vs.Document, error) {
	lcdocs, err := lcgosplitter.SplitDocuments(a.lc, types.ToLangchainDocs(docs))
	return types.FromLangchainDocs(lcdocs), err
}

func FromLangchain(splitter lcgosplitter.TextSplitter) types.TextSplitter {
	return &langchainSplitterAdapter{splitter}
}

func LangchainToNative(splitter lcgosplitter.TextSplitter) types.TextSplitter {
	return FromLangchain(splitter)
}

func AsLangchain(splitter types.TextSplitter) lcgosplitter.TextSplitter {
	return splitter.(*langchainSplitterAdapter).lc
}
