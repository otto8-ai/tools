package markdown_basic

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

	MaxHeadingLevel   int  // Maximum heading level to split on
	IgnoreHeadingOnly bool // Ignore chunks that only contain headings
}

// DefaultOptions returns the default options for all text splitter.
func DefaultOptions() Options {
	return Options{
		ChunkSize:    defaults.TextSplitterChunkSize,
		ChunkOverlap: defaults.TextSplitterChunkOverlap,

		ModelName:    defaults.TextSplitterTokenModel,
		EncodingName: defaults.TextSplitterTokenEncoding,

		MaxHeadingLevel:   6,
		IgnoreHeadingOnly: false,
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

// WithSecondSplitter sets the second splitter for a text splitter.
func WithSecondSplitter(secondSplitter lcgosplitter.TextSplitter) Option {
	return func(o *Options) {
		o.SecondSplitter = secondSplitter
	}
}

// WithKeepSeparator sets whether the separators should be kept in the resulting
// split text or not. When it is set to True, the separators are included in the
// resulting split text. When it is set to False, the separators are not included
// in the resulting split text. The purpose of having this parameter is to provide
// flexibility in how text splitting is handled. Default to False if not specified.
func WithKeepSeparator(keepSeparator bool) Option {
	return func(o *Options) {
		o.KeepSeparator = keepSeparator
	}
}

func WithMaxHeadingLevel(maxHeadingLevel int) Option {
	return func(o *Options) {
		o.MaxHeadingLevel = maxHeadingLevel
	}
}

func WithIgnoreHeadingOnly(ignoreHeadingOnly bool) Option {
	return func(o *Options) {
		o.IgnoreHeadingOnly = ignoreHeadingOnly
	}
}
