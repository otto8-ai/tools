package documentloader

import (
	"context"

	"github.com/gptscript-ai/knowledge/pkg/datastore/textsplitter"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	golcschema "github.com/hupe1980/golc/schema"
	lcgodocloaders "github.com/tmc/langchaingo/documentloaders"
)

type golcAdapter struct {
	golcschema.DocumentLoader
}

func (a *golcAdapter) Load(ctx context.Context) ([]vs.Document, error) {
	golcdocs, err := a.DocumentLoader.Load(ctx)
	return types.FromGolcDocs(golcdocs), err
}

func (a *golcAdapter) LoadAndSplit(ctx context.Context, splitter types.TextSplitter) ([]vs.Document, error) {
	golcdocs, err := a.DocumentLoader.LoadAndSplit(ctx, textsplitter.AsGolc(splitter))
	return types.FromGolcDocs(golcdocs), err
}

func FromGolc(loader golcschema.DocumentLoader) types.DocumentLoader {
	return &golcAdapter{loader}
}

func AsGolc(loader types.DocumentLoader) golcschema.DocumentLoader {
	return loader.(*golcAdapter).DocumentLoader
}

// --- langchaingo ---

type langchainAdapter struct {
	lcgodocloaders.Loader
}

func (a *langchainAdapter) Load(ctx context.Context) ([]vs.Document, error) {
	lcdocs, err := a.Loader.Load(ctx)
	return types.FromLangchainDocs(lcdocs), err
}

func (a *langchainAdapter) LoadAndSplit(ctx context.Context, splitter types.TextSplitter) ([]vs.Document, error) {
	lcdocs, err := a.Loader.LoadAndSplit(ctx, textsplitter.AsLangchain(splitter))
	return types.FromLangchainDocs(lcdocs), err
}

func FromLangchain(loader lcgodocloaders.Loader) types.DocumentLoader {
	return &langchainAdapter{loader}
}

func AsLangchain(loader types.DocumentLoader) lcgodocloaders.Loader {
	return loader.(*langchainAdapter).Loader
}
