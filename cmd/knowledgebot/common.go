package main

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/mgoltzsche/knowledgebot/internal/qdrantutils"
	"github.com/spf13/pflag"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

type LLMFactory struct {
	APIURL         string
	APIKey         string
	Model          string
	EmbeddingModel string
}

func (f *LLMFactory) AddLLMFlags(fs *pflag.FlagSet) {
	fs.StringVar(&f.APIURL, "openai-url", f.APIURL, "URL pointing to the OpenAI LLM API server")
	fs.StringVar(&f.APIKey, "openai-key", f.APIKey, "API key for the OpenAI LLM API")
	fs.StringVar(&f.Model, "model", f.Model, "LLM model to use")
}

func (f *LLMFactory) NewLLM() (*openai.LLM, error) {
	return openai.New(
		openai.WithHTTPClient(&http.Client{Timeout: 90 * time.Second}),
		openai.WithBaseURL(f.APIURL+"/v1"),
		openai.WithToken(f.APIKey),
		openai.WithModel(f.Model),
		openai.WithEmbeddingModel(f.EmbeddingModel),
	)
}

type StoreFactory struct {
	LLMFactory
	EmbeddingDimensions int
	QdrantURL           string
	QdrantCollection    string
}

func (f *StoreFactory) AddStoreFlags(fs *pflag.FlagSet) {
	fs.StringVar(&f.EmbeddingModel, "embedding-model", f.EmbeddingModel, "Embedding model dimensions")
	fs.IntVar(&f.EmbeddingDimensions, "embedding-dimensions", f.EmbeddingDimensions, "LLM embedding model to use")
	fs.StringVar(&f.QdrantURL, "qdrant-url", f.QdrantURL, "LLM model to use")
	fs.StringVar(&f.QdrantCollection, "qdrant-collection", f.QdrantCollection, "LLM model to use")
}

func (f *StoreFactory) NewStore() (vectorstores.VectorStore, error) {
	llm, err := f.NewLLM()
	if err != nil {
		return nil, err
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, err
	}

	qdrantURL, err := url.Parse(f.QdrantURL)
	if err != nil {
		return nil, err
	}

	return qdrant.New(
		qdrant.WithURL(*qdrantURL),
		qdrant.WithCollectionName(f.QdrantCollection),
		qdrant.WithEmbedder(e),
	)
}

func (f *StoreFactory) CreateCollectionIfNotExist(ctx context.Context) error {
	return qdrantutils.CreateQdrantCollectionIfNotExist(ctx, f.QdrantURL, f.QdrantCollection, f.EmbeddingDimensions)
}
