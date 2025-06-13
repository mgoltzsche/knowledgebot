package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gocolly/colly"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
)

type Crawler struct {
	MaxDepth         int
	MaxPages         uint64
	URLRegex         *regexp.Regexp
	Sink             vectorstores.VectorStore
	mutex            sync.Mutex
	knownChunkHashes map[string]struct{}
}

func (s *Crawler) Crawl(ctx context.Context, seedURL string) error {
	log.Printf("crawling %s with maxDepth=%d, maxPages=%d and urlRegex=%s", seedURL, s.MaxDepth, s.MaxPages, s.URLRegex)

	startTime := time.Now()

	u, err := url.Parse(seedURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan []schema.Document, 50)

	go s.crawl(ctx, u, ch)

	return s.indexDocumentChunks(ctx, cancel, ch, startTime)
}

func (s *Crawler) crawl(ctx context.Context, seedURL *url.URL, ch chan<- []schema.Document) error {
	defer close(ch)

	pageCounter := atomic.Uint64{}
	domain := strings.TrimPrefix(seedURL.Host, "www.")
	opts := []func(*colly.Collector){
		colly.MaxDepth(s.MaxDepth),
		colly.AllowedDomains(seedURL.Hostname(), domain),
		colly.DetectCharset(),
		colly.UserAgent("knowledgebot"),
	}

	if s.URLRegex != nil {
		opts = append(opts, colly.URLFilters(s.URLRegex))
	}

	c := colly.NewCollector(opts...)

	c.OnRequest(func(req *colly.Request) {
		select {
		case <-ctx.Done():
			req.Abort()
			return
		default:
		}

		if s.MaxPages > 0 {
			if pageCounter.Add(1) > s.MaxPages {
				req.Abort()
				return
			}
		}

		log.Println("visiting", req.URL)
	})

	c.OnResponse(func(f *colly.Response) {
		err := s.processHTML(ctx, f.Request.URL, string(f.Body), ch)
		if err != nil {
			log.Println("WARNING:", err)
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.Visit(seedURL.String())
	c.Wait()

	return ctx.Err()
}

func (s *Crawler) processHTML(ctx context.Context, url *url.URL, html string, ch chan<- []schema.Document) error {
	markdown, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return fmt.Errorf("html to markdown: %w", err)
	}

	markdown = stripMarkdownLinks(markdown)

	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(768),
		textsplitter.WithChunkOverlap(175),
	)

	chunks, err := splitter.SplitText(markdown)
	if err != nil {
		return fmt.Errorf("split text: %w", err)
	}

	log.Printf("scraped %d chunks from %s", len(chunks), url)

	docs := make([]schema.Document, 0, len(chunks))

	for _, chunk := range chunks {
		if s.knownChunk(chunk) {
			continue
		}

		docs = append(docs, schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"url":   url.String(),
				"title": deriveTitle(markdown, url),
			},
		})
	}

	if len(docs) > 0 {
		ch <- docs
	}

	return nil
}

func (s *Crawler) knownChunk(chunk string) bool {
	h := sha256.New()
	_, _ = h.Write([]byte(chunk))
	b := h.Sum(nil)
	key := hex.EncodeToString(b)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.knownChunkHashes == nil {
		s.knownChunkHashes = map[string]struct{}{}
	}

	_, known := s.knownChunkHashes[key]
	if !known {
		s.knownChunkHashes[key] = struct{}{}
	}

	return known
}

func deriveTitle(markdown string, u *url.URL) string {
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "# ") {
			title := strings.Trim(line[2:], "*_ ")
			if len(title) > 0 {
				return fmt.Sprintf("%s | %s", title, u.Hostname())
			}
		}
	}

	pathSegments := strings.Split(u.Path, "/")
	lastPathSegment := pathSegments[len(pathSegments)-1]

	if lastPathSegment == "" || lastPathSegment == "." && len(pathSegments) > 1 {
		lastPathSegment = pathSegments[len(pathSegments)-2]
		if lastPathSegment == "" || lastPathSegment == "." {
			lastPathSegment = u.Path
		}
	}

	return fmt.Sprintf("%s | %s", lastPathSegment, u.Hostname())
}

func (s *Crawler) indexDocumentChunks(ctx context.Context, cancel context.CancelFunc, ch <-chan []schema.Document, startTime time.Time) error {
	var err error

	docCount := 0
	chunkCount := 0

	for chunks := range ch {
		if err == nil {
			_, e := s.Sink.AddDocuments(ctx, chunks)
			if e != nil {
				err = e
				cancel()
			}

			docCount++
			chunkCount += len(chunks)
		}
	}

	if err != nil {
		return fmt.Errorf("index scraped chunks: %w", err)
	}

	elapsed := time.Now().Sub(startTime)

	log.Printf("indexed %d chunks of %d document(s) in %s", chunkCount, docCount, elapsed)

	return ctx.Err()
}
