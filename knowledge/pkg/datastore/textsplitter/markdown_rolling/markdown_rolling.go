package markdown_rolling

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	lcgosplitter "github.com/tmc/langchaingo/textsplitter"
)

// NewMarkdownTextSplitter creates a new Markdown text splitter.
func NewMarkdownTextSplitter(opts ...Option) (*MarkdownTextSplitter, error) {
	options := DefaultOptions()

	for _, opt := range opts {
		opt(&options)
	}

	var tk *tiktoken.Tiktoken
	var err error
	if options.EncodingName != "" {
		tk, err = tiktoken.GetEncoding(options.EncodingName)
	} else {
		tk, err = tiktoken.EncodingForModel(options.ModelName)
	}
	if err != nil {
		return nil, fmt.Errorf("couldn't get encoding: %w", err)
	}

	tokenSplitter := lcgosplitter.TokenSplitter{
		ChunkSize:         options.ChunkSize,
		ChunkOverlap:      options.ChunkOverlap,
		ModelName:         options.ModelName,
		EncodingName:      options.EncodingName,
		AllowedSpecial:    []string{},
		DisallowedSpecial: []string{"all"},
	}

	return &MarkdownTextSplitter{
		options,
		tk,
		tokenSplitter,
	}, nil
}

// MarkdownTextSplitter markdown header text splitter.
type MarkdownTextSplitter struct {
	Options
	*tiktoken.Tiktoken
	tokenSplitter lcgosplitter.TokenSplitter
}

type block struct {
	headings  []string
	lines     []string
	text      string
	tokenSize int
}

func (s *MarkdownTextSplitter) getTokenSize(text string) int {
	return len(s.Encode(text, []string{}, []string{"all"}))
}

func (s *MarkdownTextSplitter) finishBlock(blocks []block, currentBlock block, headingStack []string) ([]block, block, error) {

	for _, header := range headingStack {
		if header != "" {
			currentBlock.headings = append(currentBlock.headings, header)
		}
	}

	if (len(currentBlock.lines) == 0) && s.IgnoreHeadingOnly {
		return blocks, block{}, nil
	}

	blockHeading := strings.TrimSpace(currentB)

	headingStr := strings.TrimSpace(strings.Join(currentBlock.headings, "\n"))
	contentStr := strings.TrimSpace(strings.Join(currentBlock.lines, "\n"))
	text := headingStr + "\n" + contentStr

	if len(text) == 0 {
		return blocks, block{}, nil
	}

	textTokenSize := s.getTokenSize(text)

	if textTokenSize <= s.ChunkSize {
		// append new block to free up some space
		return append(blocks, block{
			text:      text,
			tokenSize: textTokenSize,
		}), block{}, nil
	}

	// If the block is larger than the chunk size, split it
	headingTokenSize := s.getTokenSize(headingStr)

	// Split into chunks that leave room for the heading
	s.tokenSplitter.ChunkSize = s.ChunkSize - headingTokenSize

	splits, err := s.tokenSplitter.SplitText(contentStr)
	if err != nil {
		return blocks, block{}, err
	}

	for _, split := range splits {
		text = headingStr + "\n" + split
		blocks = append(blocks, block{
			text:      text,
			tokenSize: s.getTokenSize(text),
		})
	}

	return blocks, block{}, nil

}

// SplitText splits text into chunks.
func (s *MarkdownTextSplitter) SplitText(text string) ([]string, error) {

	var (
		headingStack        []string
		chunks              []string
		currentChunk        block
		currentHeadingLevel int = 1
		currentBlock        block

		blocks []block
		err    error
	)

	// Parse markdown line-by-line and build heading-delimited blocks
	for _, line := range strings.Split(text, "\n") {

		// Handle header = start a new block
		if strings.HasPrefix(line, "#") {
			// Finish the previous Block
			blocks, currentBlock, err = s.finishBlock(blocks, currentBlock, headingStack)
			if err != nil {
				return nil, err
			}

			// Get the header level
			headingLevel := strings.Count(strings.Split(line, " ")[0], "#") - 1

			headingStack = append(headingStack[:headingLevel], line)

			// Clear the header stack for lower level headers
			for j := headingLevel + 1; j < len(headingStack); j++ {
				headingStack[j] = ""
			}

			// Reset header stack indices between this level and the last seen level, backwards
			for j := headingLevel - 1; j > currentHeadingLevel; j-- {
				headingStack[j] = ""
			}

			currentHeadingLevel = headingLevel
			continue

		}

		// If the line is not a header, add it to the current block
		currentBlock.lines = append(currentBlock.lines, line)

	}

	// Finish the last block
	blocks, currentBlock, err = s.finishBlock(blocks, currentBlock, headingStack)
	if err != nil {
		return nil, err
	}

	// Combine blocks into chunks as close to the target token size as possible
	for _, b := range blocks {
		if currentChunk.tokenSize+b.tokenSize <= s.ChunkSize {
			// Doesn't exceed chunk size, so add to the current chunk
			currentChunk.text += "\n" + b.text
			currentChunk.tokenSize += b.tokenSize
		} else {
			// Exceeds chunk size, so start a new chunk
			chunks = append(chunks, strings.TrimSpace(currentChunk.text))
			currentChunk = b
		}
	}

	// Append the last chunk if it wasn't added
	if currentChunk.text != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk.text))
	}

	return chunks, nil
}
