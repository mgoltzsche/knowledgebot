package qdrantutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func CreateQdrantCollectionIfNotExist(ctx context.Context, qdrantURL, collection string, dimensions int) error {
	body, err := json.Marshal(map[string]any{
		"vectors": map[string]any{
			"size":     dimensions,
			"distance": "Cosine",
			"datatype": "float16",
		},
	})
	if err != nil {
		return fmt.Errorf("create qdrant collection: marshal request body: %w", err)
	}

	collectionURL := fmt.Sprintf("%s/collections/%s", qdrantURL, url.PathEscape(collection))
	httpClient := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, collectionURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create qdrant collection: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("create qdrant collection: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("created qdrant collection %q", collection)
		return nil
	}

	if resp.StatusCode == http.StatusConflict {
		return nil
	}

	return fmt.Errorf("create qdrant collection: server responded with %s", resp.Status)
}
