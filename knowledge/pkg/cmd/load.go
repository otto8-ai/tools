package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/client"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/structured"
	"github.com/gptscript-ai/knowledge/pkg/datastore/filetypes"
	"github.com/spf13/cobra"
)

type ClientLoad struct {
	Loader       string            `usage:"Choose a document loader to use"`
	OutputFormat string            `name:"format" usage:"Choose an output format" default:"structured"`
	Metadata     map[string]string `usage:"Metadata to attach to the loaded files" env:"METADATA"`
	MetadataJSON string            `usage:"Metadata to attach to the loaded files in JSON format" env:"METADATA_JSON"`
}

func (s *ClientLoad) Customize(cmd *cobra.Command) {
	cmd.Use = "load <input> <output>"
	cmd.Short = "Load a file and transform it to markdown"
	cmd.Args = cobra.ExactArgs(2)
}

func (s *ClientLoad) Run(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := args[1]

	err := s.run(cmd.Context(), input, output)
	if err != nil {
		exitErr0(err)
	}
	return nil
}

func (s *ClientLoad) run(ctx context.Context, input, output string) error {
	if !slices.Contains([]string{"structured", "markdown"}, s.OutputFormat) {
		return fmt.Errorf("unsupported output format %q", s.OutputFormat)
	}

	var metadata map[string]string
	if s.MetadataJSON != "" {
		if err := json.Unmarshal([]byte(s.MetadataJSON), &metadata); err != nil {
			return fmt.Errorf("failed to unmarshal metadata JSON: %w", err)
		}
	}
	maps.Copy(metadata, s.Metadata)

	c, err := client.NewStandaloneClient(ctx, nil)
	if err != nil {
		return err
	}

	var inputBytes []byte
	if strings.HasPrefix(input, "ws://") {
		inputBytes, err = c.GPTScript.ReadFileInWorkspace(ctx, strings.TrimPrefix(input, "ws://"))
	} else {
		inputBytes, err = os.ReadFile(input)
	}
	if err != nil {
		return fmt.Errorf("failed to read input file %q: %w", input, err)
	}

	filetype, err := filetypes.GetFiletype(input, inputBytes)
	if err != nil {
		return fmt.Errorf("failed to get filetype for input file %q: %w", input, err)
	}

	var loader documentloader.LoaderFunc

	if s.Loader == "" {
		loader = documentloader.DefaultDocLoaderFunc(filetype, documentloader.DefaultDocLoaderFuncOpts{})
	} else {
		var err error
		loader, err = documentloader.GetDocumentLoaderFunc(s.Loader, nil)
		if err != nil {
			return fmt.Errorf("failed to get document loader function %q: %w", s.Loader, err)
		}
	}

	if loader == nil {
		return fmt.Errorf("unsupported file type %q", input)
	}

	docs, err := loader(ctx, bytes.NewReader(inputBytes))
	if err != nil {
		return fmt.Errorf("failed to load documents from file %q using loader %q: %w", input, s.Loader, err)
	}

	var text string

	switch s.OutputFormat {
	case "markdown":
		var texts []string
		for _, doc := range docs {
			if len(doc.Content) == 0 {
				continue
			}

			for k, v := range metadata {
				doc.Metadata[k] = v
			}

			metadata, err := json.Marshal(doc.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}

			content := fmt.Sprintf("!metadata %s\n%s", metadata, doc.Content)

			texts = append(texts, content)
		}

		text = strings.Join(texts, "\n---docbreak---\n")

	case "structured":
		var structuredInput structured.StructuredInput
		structuredInput.Metadata = map[string]any{}
		structuredInput.Documents = make([]structured.StructuredInputDocument, 0, len(docs))

		commonMetadata := maps.Clone(docs[0].Metadata)
		for _, doc := range docs {
			commonMetadata = extractCommon(commonMetadata, doc.Metadata)
			structuredInput.Documents = append(structuredInput.Documents, structured.StructuredInputDocument{
				Metadata: doc.Metadata,
				Content:  doc.Content,
			})
		}

		commonMetadata["source"] = input

		for k, v := range metadata {
			commonMetadata[k] = v
		}
		structuredInput.Metadata = commonMetadata

		for i, doc := range structuredInput.Documents {
			structuredInput.Documents[i].Metadata = dropCommon(doc.Metadata, commonMetadata)
		}

		jsonBytes := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(jsonBytes)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(structuredInput); err != nil {
			return fmt.Errorf("failed to encode structured input: %w", err)
		}
		text = jsonBytes.String()
	default:
		return fmt.Errorf("unsupported output format %q", s.OutputFormat)
	}

	if output == "-" {
		fmt.Println(text)
		return nil
	}

	if strings.HasPrefix(output, "ws://") {
		return c.GPTScript.WriteFileInWorkspace(ctx, strings.TrimPrefix(output, "ws://"), []byte(text))
	}

	return os.WriteFile(output, []byte(text), 0666)
}

func dropCommon(target, common map[string]any) map[string]any {
	for key, _ := range target {
		if _, exists := common[key]; exists {
			delete(target, key)
		}
	}

	return target
}

func extractCommon(target, other map[string]any) map[string]any {
	for key, value := range target {
		if v, exists := other[key]; exists && v == value {
			target[key] = value
		} else {
			delete(target, key)
		}
	}

	return target
}
