package qna

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type QuestionAnswerWorkflow struct {
	LLM            llms.LLM
	Temperature    float64
	Store          vectorstores.VectorStore
	MaxDocs        int
	ScoreThreshold float64
}

type ResponseChunk struct {
	Err     error             `json:"error,omitempty"`
	Chunk   string            `json:"chunk,omitempty"`
	Sources []SourceReference `json:"sources,omitempty"`
}

type SourceReference struct {
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	MaxScore float32   `json:"maxScore"`
	Snippets []Snippet `json:"snippets,omitempty"`
}

type Snippet struct {
	Text  string  `json:"text"`
	Score float32 `json:"score"`
}

func (w *QuestionAnswerWorkflow) Answer(ctx context.Context, question string) (<-chan ResponseChunk, error) {
	docs, err := w.Store.SimilaritySearch(ctx, question, w.MaxDocs, vectorstores.WithScoreThreshold(float32(w.ScoreThreshold)))
	if err != nil {
		return nil, fmt.Errorf("query knowledge base: %w", err)
	}

	sourceRefs := searchResultsToSourceRefs(docs)

	ch := make(chan ResponseChunk)
	prompt := buildPrompt(docs)

	log.Println("Requesting LLM answer for prompt:\n  ", strings.ReplaceAll(prompt, "\n", "\n  "))

	go func() {
		defer close(ch)

		if len(sourceRefs) > 0 {
			ch <- ResponseChunk{Sources: sourceRefs}
		}

		_, err := w.LLM.GenerateContent(ctx,
			[]llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeSystem, prompt),
				llms.TextParts(llms.ChatMessageTypeHuman, question),
			},
			llms.WithStreamingFunc(w.streamFunc(ch)),
			llms.WithTemperature(w.Temperature),
		)
		if err != nil && !errors.Is(err, context.Canceled) {
			ch <- ResponseChunk{Err: err}
		}
	}()

	return ch, nil
}

func (w *QuestionAnswerWorkflow) streamFunc(ch chan<- ResponseChunk) func(ctx context.Context, chunk []byte) error {
	return func(ctx context.Context, chunk []byte) error {
		if len(chunk) > 0 {
			ch <- ResponseChunk{Chunk: string(chunk)}
		}

		return nil
	}
}

func searchResultsToSourceRefs(docs []schema.Document) []SourceReference {
	urlMap := make(map[string]*SourceReference, len(docs))
	urls := make([]string, 0, len(docs))

	for _, doc := range docs {
		urlKey, ok := doc.Metadata["url"].(string)
		if !ok {
			log.Println("WARNING: vectordb search result doc does not specify 'url' metadata key")
			continue
		}

		title, ok := doc.Metadata["title"].(string)
		if !ok {
			log.Println("WARNING: vectordb search result doc does not specify 'title' metadata key")
			continue
		}

		ref, ok := urlMap[urlKey]
		if !ok {
			ref = &SourceReference{
				URL:      urlKey,
				Title:    title,
				Snippets: make([]Snippet, 0, 1),
			}
			urlMap[urlKey] = ref
			urls = append(urls, urlKey)
		}

		ref.Snippets = append(ref.Snippets, Snippet{
			Text:  doc.PageContent,
			Score: doc.Score,
		})

		if doc.Score > ref.MaxScore {
			ref.MaxScore = doc.Score
		}
	}

	refs := make([]SourceReference, len(urls))
	for i, key := range urls {
		refs[i] = *urlMap[key]
	}

	return refs
}

func buildPrompt(docs []schema.Document) string {
	related := make([]string, len(docs))
	for i, doc := range docs {
		related[i] = doc.PageContent
	}

	return fmt.Sprintf("You are a helpful assistant.\n\nAnswer the user's questions short and concise based on the following information:%s\n\n.", strings.Join(related, "\n\n"))
}
