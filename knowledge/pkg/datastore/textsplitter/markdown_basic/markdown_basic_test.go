package markdown_basic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitTextWithBasicMarkdown(t *testing.T) {
	splitter := NewMarkdownTextSplitter()
	chunks, err := splitter.SplitText("# Heading\n\nThis is a paragraph.")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	expected := []string{"# Heading\nThis is a paragraph."}

	assert.Equal(t, expected, chunks)
}

func TestSplitTextWithOptions(t *testing.T) {
	md := `
# Heading 1

some p under h1

## Heading 2
### Heading 3

- some
- list
- items

**bold**

# 2nd Heading 1
#### Heading 4

some p under h4
`

	testcases := []struct {
		name     string
		splitter *MarkdownTextSplitter
		expected []string
	}{
		{
			name:     "default",
			splitter: NewMarkdownTextSplitter(),
			expected: []string{
				"# Heading 1\nsome p under h1",
				"# Heading 1\n## Heading 2",
				"# Heading 1\n## Heading 2\n### Heading 3\n- some\n- list\n- items\n\n**bold**",
				"# 2nd Heading 1",
				"# 2nd Heading 1\n#### Heading 4\nsome p under h4",
			},
		},
		{
			name:     "ignore_heading_only",
			splitter: NewMarkdownTextSplitter(WithIgnoreHeadingOnly(true)),
			expected: []string{
				"# Heading 1\nsome p under h1",
				"# Heading 1\n## Heading 2\n### Heading 3\n- some\n- list\n- items\n\n**bold**",
				"# 2nd Heading 1\n#### Heading 4\nsome p under h4",
			},
		},
		{
			name:     "split_h1_only",
			splitter: NewMarkdownTextSplitter(WithMaxHeadingLevel(1)),
			expected: []string{
				"# Heading 1\nsome p under h1\n\n## Heading 2\n### Heading 3\n\n- some\n- list\n- items\n\n**bold**",
				"# 2nd Heading 1\n#### Heading 4\n\nsome p under h4",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := tc.splitter.SplitText(md)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), len(chunks))

			assert.Equal(t, tc.expected, chunks)
		})
	}
}
