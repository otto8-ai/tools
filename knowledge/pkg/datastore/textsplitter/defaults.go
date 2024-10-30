package textsplitter

import (
	"github.com/acorn-io/z"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
)

func DefaultTextSplitter(filetype string, textSplitterOpts *TextSplitterOpts) types.TextSplitter {
	if textSplitterOpts == nil {
		textSplitterOpts = z.Pointer(NewTextSplitterOpts())
	}
	genericTextSplitter := FromLangchain(NewLcgoTextSplitter(*textSplitterOpts))
	markdownTextSplitter := FromLangchain(NewLcgoMarkdownSplitter(*textSplitterOpts))

	switch filetype {
	case ".md", "text/markdown":
		return markdownTextSplitter
	default:
		return genericTextSplitter
	}
}
