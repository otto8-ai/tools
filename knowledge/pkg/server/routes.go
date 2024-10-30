package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/acorn-io/z"
	"github.com/gin-gonic/gin"
	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/gptscript-ai/knowledge/pkg/index"
	"github.com/gptscript-ai/knowledge/pkg/server/types"
)

// CreateDS creates a new dataset.
// @Summary Create a new dataset
// @Description Create a new dataset
// @Tags datasets
// @Accept json
// @Produce json
// @Param dataset body types.Dataset true "Dataset object"
// @Router /datasets/create [post]
// @Success 200 {object} types.Dataset
func (s *Server) CreateDS(c *gin.Context) {
	var dataset types.Dataset
	if err := c.ShouldBindJSON(&dataset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create Dataset
	if err := s.NewDataset(c, index.Dataset{ID: dataset.ID}); err != nil {
		slog.Error("Failed to create dataset", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataset)
}

// DeleteDS deletes a dataset by ID.
// @Summary Delete a dataset
// @Description Delete a dataset by ID
// @Tags datasets
// @Produce json
// @Param id path string true "Dataset ID"
// @Router /datasets/{id} [delete]
// @Success 200 {object} gin.H
func (s *Server) DeleteDS(c *gin.Context) {
	id := c.Param("id")
	slog.Info("Deleting dataset", "id", id)

	if err := s.DeleteDataset(c, id); err != nil {
		slog.Error("Failed to delete dataset", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// RetrieveFromDS retrieves content from a dataset by ID.
// @Summary Retrieve content from a dataset
// @Description Retrieve content from a dataset by ID
// @Tags datasets
// @Accept json
// @Produce json
// @Param id path string true "Dataset ID"
// @Router /datasets/{id}/retrieve [post]
// @Success 200 {object} []vectorstore.Document
func (s *Server) RetrieveFromDS(c *gin.Context) {
	id := c.Param("id")
	slog.Info("Retrieving content from dataset", "dataset", id)

	var query types.Query
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: support retrieval flows and keywords
	docs, err := s.Retrieve(c, []string{id}, query.Prompt, datastore.RetrieveOpts{TopK: z.Dereference(query.TopK)})
	if err != nil {
		slog.Error("Failed to retrieve documents", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, docs)
}

// IngestIntoDS ingests content into a dataset by ID.
// @Summary Ingest content into a dataset
// @Description Ingest content into a dataset by ID
// @Tags datasets
// @Accept json
// @Produce json
// @Param id path string true "Dataset ID"
// @Router /datasets/{id}/ingest [post]
// @Success 200 {object} types.IngestResponse
func (s *Server) IngestIntoDS(c *gin.Context) {
	id := c.Param("id")
	slog.Info("Ingesting content into dataset", "dataset", id)

	var ingest types.Ingest
	if err := c.ShouldBindJSON(&ingest); err != nil {
		slog.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Debug("Received ingest request", "content_size", len(ingest.Content), "metadata", ingest.FileMetadata)

	// decode content
	data, err := base64.StdEncoding.DecodeString(ingest.Content)
	if err != nil {
		slog.Error("Failed to decode content", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ingest content
	// TODO: support ingestion flows
	docIDs, err := s.Ingest(c, id, z.Dereference(ingest.Filename), data, datastore.IngestOpts{
		FileMetadata:     ingest.FileMetadata,
		TextSplitterOpts: ingest.TextSplitterOpts,
	})

	if err != nil {
		slog.Error("Failed to ingest content", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.IngestResponse{Documents: docIDs})
}

// RemoveDocFromDS removes a document from a dataset by ID. If the owning file context is now empty, the FileIndex is removed.
// @Summary Remove a document from a dataset
// @Description Remove a document from a dataset by ID
// @Tags datasets
// @Accept json
// @Produce json
// @Param id path string true "Dataset ID"
// @Param doc_id path string true "Document ID"
// @Router /datasets/{id}/documents/{doc_id} [delete]
// @Success 200 {object} gin.H
func (s *Server) RemoveDocFromDS(c *gin.Context) {
	id := c.Param("id")
	docID := c.Param("doc_id")
	slog.Info("Removing document from dataset", "dataset", id, "document", docID)

	if err := s.DeleteDocument(c, id, docID); err != nil {
		if errors.Is(err, datastore.ErrDBDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"msg": err.Error()})
			return
		}
		slog.Error("Failed to remove document", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "doc_id": docID})
}

// RemoveFileFromDS removes a file from a dataset by ID.
// @Summary Remove a file from a dataset
// @Description Remove a file from a dataset by ID
// @Tags datasets
// @Accept json
// @Produce json
// @Param id path string true "Dataset ID"
// @Param file_id path string true "File ID"
// @Router /datasets/{id}/files/{file_id} [delete]
// @Success 200 {object} gin.H
func (s *Server) RemoveFileFromDS(c *gin.Context) {
	id := c.Param("id")
	fileID := c.Param("file_id")
	slog.Info("Removing file from dataset", "dataset", id, "file", fileID)

	err := s.DeleteFile(c, id, fileID)
	if err != nil {
		if errors.Is(err, datastore.ErrDBFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"msg": err.Error()})
			return
		}
		slog.Error("Failed to remove file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "file_id": fileID})
}

// ListDS lists all datasets.
// @Summary List all datasets
// @Description List all datasets
// @Tags datasets
// @Produce json
// @Router /datasets [get]
// @Success 200 {object} []types.Dataset
func (s *Server) ListDS(c *gin.Context) {
	slog.Info("Listing datasets")

	datasets, err := s.ListDatasets(c)
	if err != nil {
		slog.Error("Failed to list datasets", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	datasetsResponse := make([]types.Dataset, len(datasets))
	for i, dataset := range datasets {
		datasetsResponse[i] = types.Dataset{
			ID: dataset.ID,
		}
	}

	c.JSON(http.StatusOK, datasetsResponse)
}

// GetDS gets a dataset by ID.
// @Summary Get a dataset
// @Description Get a dataset by ID
// @Tags datasets
// @Produce json
// @Param id path string true "Dataset ID"
// @Router /datasets/{id} [get]
// @Success 200 {object} types.Dataset
func (s *Server) GetDS(c *gin.Context) {
	id := c.Param("id")
	slog.Info("Getting dataset", "id", id)

	dataset, err := s.GetDataset(c, id)
	if err != nil {
		slog.Error("Failed to get dataset", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if dataset == nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("dataset not found: %v", id), "id": id})
		return
	}

	slog.Info("Found dataset", "id", dataset.ID, "num_files", len(dataset.Files))

	c.JSON(http.StatusOK, dataset)
}
