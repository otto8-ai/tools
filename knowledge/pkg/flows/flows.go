package flows

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"time"

	"github.com/acorn-io/z"
	"github.com/google/uuid"
	"github.com/gptscript-ai/knowledge/pkg/datastore/store"
	"github.com/gptscript-ai/knowledge/pkg/log"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/mitchellh/mapstructure"
	"github.com/philippgille/chromem-go"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader"
	"github.com/gptscript-ai/knowledge/pkg/datastore/postprocessors"
	"github.com/gptscript-ai/knowledge/pkg/datastore/querymodifiers"
	"github.com/gptscript-ai/knowledge/pkg/datastore/retrievers"
	"github.com/gptscript-ai/knowledge/pkg/datastore/textsplitter"
	"github.com/gptscript-ai/knowledge/pkg/datastore/transformers"
	dstypes "github.com/gptscript-ai/knowledge/pkg/datastore/types"
)

type IngestionFlowGlobals struct {
	SplitterOpts map[string]any
}

type IngestionFlow struct {
	Globals         IngestionFlowGlobals
	Filetypes       []string
	Load            documentloader.LoaderFunc
	Splitter        dstypes.TextSplitter
	Transformations []dstypes.DocumentTransformer
}

func (f *IngestionFlow) Transform(ctx context.Context, docs []vs.Document) ([]vs.Document, error) {
	var err error
	for i, t := range f.Transformations {
		l := log.FromCtx(ctx).With("transformer", t.Name()).With("progress", fmt.Sprintf("%d/%d", i+1, len(f.Transformations))).With("progress_unit", "transformations")
		l.Info("Running transformer")
		docs, err = t.Transform(ctx, docs)
		if err != nil {
			l.With("status", "failed").Error("Failed to transform documents", "error", err)
			return nil, err
		}
		l.With("status", "completed").Info("Transformed documents", "num_documents", len(docs))
	}
	return docs, nil
}

func NewDefaultIngestionFlow(filetype string, textsplitterOpts *textsplitter.TextSplitterOpts) (IngestionFlow, error) {
	ingestionFlow := IngestionFlow{
		Filetypes: []string{filetype},
	}
	if err := ingestionFlow.FillDefaults(filetype, textsplitterOpts); err != nil {
		return IngestionFlow{}, err
	}
	return ingestionFlow, nil
}

func (f *IngestionFlow) SupportsFiletype(filetype string) bool {
	return slices.Contains(f.Filetypes, filetype) || slices.Contains(f.Filetypes, "*")
}

func (f *IngestionFlow) FillDefaults(filetype string, textsplitterOpts *textsplitter.TextSplitterOpts) error {
	if f.Load == nil {
		f.Load = documentloader.DefaultDocLoaderFunc(filetype, documentloader.DefaultDocLoaderFuncOpts{Archive: documentloader.ArchiveOpts{
			ErrOnUnsupportedFiletype: false,
			ErrOnFailedFile:          false,
		}})
	}
	if f.Splitter == nil {
		if textsplitterOpts == nil {
			textsplitterOpts = z.Pointer(textsplitter.NewTextSplitterOpts())
		}
		slog.Debug("Using default text splitter", "filetype", filetype, "textSplitterOpts", textsplitterOpts)

		if len(f.Globals.SplitterOpts) > 0 {
			if err := mapstructure.Decode(f.Globals.SplitterOpts, textsplitterOpts); err != nil {
				return fmt.Errorf("failed to decode globals.SplitterOpts configuration: %w", err)
			}
			slog.Debug("Overriding text splitter options with globals from flows config", "filetype", filetype, "textSplitterOpts", textsplitterOpts)
		}

		f.Splitter = textsplitter.DefaultTextSplitter(filetype, textsplitterOpts)
	}
	if len(f.Transformations) == 0 {
		f.Transformations = transformers.DefaultDocumentTransformers(filetype)
	}
	return nil
}

func (f *IngestionFlow) Run(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
	var err error
	var docs []vs.Document

	phaseLog := log.FromCtx(ctx).With("phase", "parse")

	/*
	 * Load documents from the content
	 * For now, we're using documentloaders from both langchaingo and golc
	 * and translate them to our document schema.
	 */

	loaderLog := phaseLog.With("stage", "documentloader")
	loaderLog.With("status", "starting").Info("Starting document loader")
	if f.Load == nil {
		loaderLog.With("status", "skipped").With("reason", "missing documentloader").Info("No documentloader available")
		return nil, nil
	}

	docs, err = f.Load(ctx, reader)
	if err != nil {
		loaderLog.With("status", "failed").Error("Failed to load documents", "error", err)
		return nil, fmt.Errorf("failed to load documents: %w", err)
	}
	loaderLog.With("status", "completed").Info("Loaded documents", "num_documents", len(docs))

	/*
	 * Split documents - Chunking
	 */
	splitterLog := phaseLog.With("stage", "textsplitter").With(slog.Int("num_documents", len(docs)))
	splitterLog.With("status", "starting").Info("Starting text splitter")

	docs, err = f.Splitter.SplitDocuments(docs)
	if err != nil {
		splitterLog.With("status", "failed").Error("Failed to split documents", "error", err)
		return nil, fmt.Errorf("failed to split documents: %w", err)
	}
	splitterLog.With("status", "completed").Info("Split documents", "new_num_documents", len(docs))

	/*
	 * Transform documents
	 */
	transformerLog := phaseLog.With("stage", "transformer").With(slog.Int("num_documents", len(docs))).With(slog.Int("num_transformers", len(f.Transformations)))
	transformerLog.With("status", "starting").Info("Starting document transformers")
	docs, err = f.Transform(ctx, docs)
	if err != nil {
		transformerLog.With("progress", "failed").Error("Failed to transform documents", "error", err)
		return nil, fmt.Errorf("failed to transform documents: %w", err)
	}
	transformerLog.With("status", "completed").Info("Transformed documents", "new_num_documents", len(docs))

	for i, doc := range docs {
		if doc.ID == "" {
			docs[i].ID = uuid.NewString()
		}
	}

	return docs, nil
}

type RetrievalFlow struct {
	QueryModifiers []querymodifiers.QueryModifier
	Retriever      retrievers.Retriever
	Postprocessors []postprocessors.Postprocessor
}

func (f *RetrievalFlow) FillDefaults(topK int) {
	if f.Retriever == nil {
		slog.Debug("No retriever specified, using basic retriever")
		f.Retriever = &retrievers.BasicRetriever{TopK: topK}
	}
}

type RetrievalFlowOpts struct {
	Where         map[string]string
	WhereDocument []chromem.WhereDocument
}

func (f *RetrievalFlow) Run(ctx context.Context, store store.Store, query string, datasetIDs []string, opts *RetrievalFlowOpts) (*dstypes.RetrievalResponse, error) {
	retrievalFlowStartTime := time.Now()

	if opts == nil {
		opts = &RetrievalFlowOpts{}
	}

	queries := []string{query}
	for _, m := range f.QueryModifiers {
		mq, err := m.ModifyQueries(queries)
		if err != nil {
			return nil, fmt.Errorf("failed to modify queries %v with QueryModifier %q: %w", queries, m.Name(), err)
		}
		slog.Debug("Modified queries", "before", queries, "queryModifier", m.Name(), "after", mq)
		queries = mq
	}
	slog.Debug("Updated query set", "query", query, "modified_query_set", queries, "num_queries", len(queries))

	response := &dstypes.RetrievalResponse{
		Query:     query,
		Datasets:  datasetIDs,
		Responses: make([]dstypes.Response, len(queries)),
	}
	for i, q := range queries {
		docs, err := f.Retriever.Retrieve(ctx, store, q, datasetIDs, opts.Where, opts.WhereDocument)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve documents for query %q using retriever %q: %w", q, f.Retriever.Name(), err)
		}
		slog.Debug("Retrieved documents", "num_documents", len(docs), "query", q, "datasets", datasetIDs, "retriever", f.Retriever.Name())
		response.Responses[i] = dstypes.Response{
			Query:           q,
			NumDocs:         len(docs),
			ResultDocuments: docs,
		}
	}

	for _, pp := range f.Postprocessors {
		err := pp.Transform(ctx, response)
		if err != nil {
			return nil, fmt.Errorf("failed to postprocess retrieval response with Postprocessor %q: %w", pp.Name(), err)
		}
	}
	slog.Debug("Postprocessed RetrievalResponse", "num_responses", len(response.Responses), "original_query", query)

	response.Stats = dstypes.Stats{
		RetrievalTimeSeconds: time.Since(retrievalFlowStartTime).Seconds(),
	}

	return response, nil
}
