package markdown_rolling

import (
	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	lcgosplitter "github.com/tmc/langchaingo/textsplitter"
)

// Options is a struct that contains options for a text splitter.
type Options struct {
	ChunkSize      int
	ChunkOverlap   int
	Separators     []string
	KeepSeparator  bool
	ModelName      string
	EncodingName   string
	SecondSplitter lcgosplitter.TextSplitter

	IgnoreHeadingOnly bool // Ignore chunks that only contain headings
}

// DefaultOptions returns the default options for all text splitter.
func DefaultOptions() Options {
	return Options{
		ChunkSize:    defaults.TextSplitterChunkSize,
		ChunkOverlap: defaults.TextSplitterChunkOverlap,

		ModelName:    defaults.TextSplitterTokenModel,
		EncodingName: defaults.TextSplitterTokenEncoding,

		IgnoreHeadingOnly: true,
	}
}

// Option is a function that can be used to set options for a text splitter.
type Option func(*Options)

// WithChunkSize sets the chunk size for a text splitter.
func WithChunkSize(chunkSize int) Option {
	return func(o *Options) {
		o.ChunkSize = chunkSize
	}
}

// WithChunkOverlap sets the chunk overlap for a text splitter.
func WithChunkOverlap(chunkOverlap int) Option {
	return func(o *Options) {
		o.ChunkOverlap = chunkOverlap
	}
}

// WithModelName sets the model name for a text splitter.
func WithModelName(modelName string) Option {
	return func(o *Options) {
		o.ModelName = modelName
	}
}

// WithEncodingName sets the encoding name for a text splitter.
func WithEncodingName(encodingName string) Option {
	return func(o *Options) {
		o.EncodingName = encodingName
	}
}

func WithIgnoreHeadingOnly(ignoreHeadingOnly bool) Option {
	return func(o *Options) {
		o.IgnoreHeadingOnly = ignoreHeadingOnly
	}
}
