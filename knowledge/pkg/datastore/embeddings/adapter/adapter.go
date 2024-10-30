package adapter

import (
	golcschema "github.com/hupe1980/golc/schema"
	cg "github.com/philippgille/chromem-go"
)

type GolcAdapter struct {
	golcschema.Embedder
}

func NewGolcAdapter(emb golcschema.Embedder) *GolcAdapter {
	return &GolcAdapter{emb}
}

func (a *GolcAdapter) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	return a.Embedder.EmbedText, nil
}
