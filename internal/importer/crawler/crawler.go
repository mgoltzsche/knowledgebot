package crawler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gocolly/colly"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
)

type Crawler struct {
	MaxDepth int
	URLRegex *regexp.Regexp
	Sink     vectorstores.VectorStore
}

func (s *Crawler) Crawl(ctx context.Context, seedURL string) error {
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

		log.Println("visiting", req.URL)
	})

	c.OnResponse(func(f *colly.Response) {
		err := processHTML(ctx, f.Request.URL, string(f.Body), ch)
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

func processHTML(ctx context.Context, url *url.URL, html string, ch chan<- []schema.Document) error {
	markdown, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return fmt.Errorf("html to markdown: %w", err)
	}

	splitter := textsplitter.NewMarkdownTextSplitter()

	chunks, err := splitter.SplitText(markdown)
	if err != nil {
		return fmt.Errorf("split text: %w", err)
	}

	log.Printf("scraped %d chunks from %s", len(chunks), url)

	if len(chunks) > 0 {
		docs := make([]schema.Document, len(chunks))

		for i, chunk := range chunks {
			docs[i] = schema.Document{
				PageContent: chunk,
				Metadata: map[string]any{
					// TODO: deduplicate URLs:
					// * save final redirect location instead of redirect source URL (to get rid of single-character URLs pointing to list of all Futurama characters URL).
					// * ignore URLs that don't follow the scheme https://en.wikipedia.org/wiki/<ARTICLE>
					// * (maybe keep track of content hashes and ignore duplicate contents)
					"url":   url.String(),
					"title": deriveTitle(markdown, url),
				},
			}
		}

		ch <- docs
	}

	return nil
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
