package textsplitter

import (
	"github.com/acorn-io/z"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
)

func DefaultTextSplitter(filetype string, textSplitterOpts *TextSplitterOpts) types.TextSplitter {
	if textSplitterOpts == nil {
		textSplitterOpts = z.Pointer(NewTextSplitterOpts())
	}
	genericTextSplitter := FromLangchain(NewLcgoTextSplitter(*textSplitterOpts), "lcgo_text")
	markdownTextSplitter := FromLangchain(NewLcgoMarkdownSplitter(*textSplitterOpts), "lcgo_markdown")

	switch filetype {
	case ".md", "text/markdown":
		return markdownTextSplitter
	default:
		return genericTextSplitter
	}
}
