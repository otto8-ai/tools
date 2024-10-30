package markdown_rolling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitTextWithBasicMarkdown(t *testing.T) {
	splitter, err := NewMarkdownTextSplitter()
	require.NoError(t, err)
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

	nts, err := NewMarkdownTextSplitter()
	require.NoError(t, err)

	nts1, err := NewMarkdownTextSplitter(WithChunkSize(8))
	require.NoError(t, err)

	testcases := []struct {
		name     string
		splitter *MarkdownTextSplitter
		expected []string
	}{
		{
			name:     "default",
			splitter: nts,
			expected: []string{
				"# Heading 1\nsome p under h1\n## Heading 2\n### Heading 3\n- some\n- list\n- items\n\n**bold**\n# 2nd Heading 1\n#### Heading 4\nsome p under h4",
			},
		},
		{
			name:     "chunk_size_8",
			splitter: nts1,
			expected: []string{
				"# Heading 1\nsome p under h1",
				"# Heading 1\n## Heading 2\n### Heading 3\n- some\n- list\n- items\n\n**bold**",
				"# 2nd Heading 1\n#### Heading 4\nsome p under h4",
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
