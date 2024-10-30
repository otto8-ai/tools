package markdown_basic

import (
	lcgosplitter "github.com/tmc/langchaingo/textsplitter"
	"strings"
	"unicode/utf8"
)

// NewMarkdownTextSplitter creates a new Markdown text splitter.
func NewMarkdownTextSplitter(opts ...Option) *MarkdownTextSplitter {
	options := DefaultOptions()

	for _, opt := range opts {
		opt(&options)
	}

	sp := &MarkdownTextSplitter{
		ChunkSize:         options.ChunkSize,
		ChunkOverlap:      options.ChunkOverlap,
		SecondSplitter:    options.SecondSplitter,
		MaxHeadingLevel:   options.MaxHeadingLevel,
		IgnoreHeadingOnly: options.IgnoreHeadingOnly,
	}

	if sp.MaxHeadingLevel == 0 {
		sp.MaxHeadingLevel = 6
	}

	return sp
}

// MarkdownTextSplitter markdown header text splitter.
type MarkdownTextSplitter struct {
	ChunkSize    int
	ChunkOverlap int
	// SecondSplitter splits paragraphs
	SecondSplitter lcgosplitter.TextSplitter

	MaxHeadingLevel int

	IgnoreHeadingOnly bool
}

// SplitText splits a text into multiple text.
func (sp MarkdownTextSplitter) SplitText(text string) ([]string, error) {
	// Parse markdown line-by-line
	headerStack := make([]string, sp.MaxHeadingLevel)
	chunks := []string{}
	currentHeaderLevel := 1
	var currentChunk []string
	var err error

	for _, line := range strings.Split(text, "\n") {
		// Handle headers: maintian a header stack
		if strings.HasPrefix(line, "#") {
			// Get the header level
			headerLevel := strings.Count(strings.Split(line, " ")[0], "#") - 1

			// If the header level is less than or equal to the max heading level
			if headerLevel < sp.MaxHeadingLevel {
				headerStack = append(headerStack[:headerLevel], line)

				// Clear the header stack for lower level headers
				for j := headerLevel + 1; j < len(headerStack); j++ {
					headerStack[j] = ""
				}

				// Reset header stack indices between this level and the last seen level, backwards
				for j := headerLevel - 1; j > currentHeaderLevel; j-- {
					headerStack[j] = ""
				}

				// If the current chunk is not empty, add it to the chunks
				chunks, currentChunk, err = sp.flushChunk(chunks, currentChunk)
				if err != nil {
					return nil, err
				}
				for _, header := range headerStack {
					if header != "" {
						currentChunk = append(currentChunk, header)
					}
				}

				currentHeaderLevel = headerLevel
				continue
			}
		}

		// If the line is not a header, add it to the current chunk
		currentChunk = append(currentChunk, line)
	}

	chunks, _, err = sp.flushChunk(chunks, currentChunk)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}

func (sp MarkdownTextSplitter) flushChunk(chunks []string, currentChunk []string) ([]string, []string, error) {
	// Ignore heading only chunks if the option is set and the last line in the chunk is a header
	if sp.IgnoreHeadingOnly && strings.HasPrefix(currentChunk[len(currentChunk)-1], "#") {
		return chunks, []string{}, nil
	}

	headings := []string{}
	for i, line := range currentChunk {
		if i >= sp.MaxHeadingLevel {
			break
		}

		if strings.HasPrefix(line, "#") {
			headings = append(headings, line)
			continue
		}
		break
	}
	headerstr := strings.Join(headings, "\n")
	contentstr := strings.Trim(strings.Join(currentChunk[len(headings):], "\n"), "\n")
	chunkstr := headerstr + "\n" + contentstr

	if chunkstr != "" && chunkstr != "\n" {
		if sp.SecondSplitter != nil {
			// Split the chunk into smaller chunks
			splits, err := sp.SecondSplitter.SplitText(chunkstr)
			if err != nil {
				return chunks, []string{}, err
			}

			chunks = append(chunks, splits...)

			if len(splits) == 0 {
				chunks = append(chunks, headerstr)
			}
		} else {
			splits, err := lcgosplitter.NewRecursiveCharacter(
				lcgosplitter.WithChunkSize(sp.ChunkSize-utf8.RuneCountInString(headerstr)),
				lcgosplitter.WithChunkOverlap(sp.ChunkOverlap),
				lcgosplitter.WithSeparators([]string{"\n\n", "\n", " ", ""}),
			).SplitText(contentstr)
			if err != nil {
				return chunks, []string{}, err
			}

			for _, split := range splits {
				chunks = append(chunks, headerstr+"\n"+split)
			}

			// headings only
			if len(splits) == 0 {
				chunks = append(chunks, headerstr)
			}
		}
	}

	return chunks, []string{}, nil
}
