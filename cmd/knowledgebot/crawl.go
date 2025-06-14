package main

import (
	"regexp"

	"github.com/mgoltzsche/knowledgebot/internal/importer/crawler"
	"github.com/spf13/cobra"
)

var (
	crawlCmd = &cobra.Command{
		Use:     "crawl URL",
		Short:   "Crawl a given website",
		Long:    `Crawl a given website.`,
		RunE:    crawlWebsite,
		PreRunE: preRunCrawl,
		Args:    cobra.ExactArgs(1),
	}
	crawl = crawler.Crawler{
		MaxDepth:     1,
		ChunkSize:    768,
		ChunkOverlap: 175,
	}
)

func init() {
	f := crawlCmd.Flags()

	f.IntVar(&crawl.MaxDepth, "max-depth", crawl.MaxDepth, "Maximum crawl depth")
	f.Uint64Var(&crawl.MaxPages, "max-pages", crawl.MaxPages, "Maximum amount of pages to crawl")
	f.Var((*urlRegexFlag)(&crawl), "url-regex", "regex to filter URLs to crawl")
	f.IntVar(&crawl.ChunkSize, "chunk-size", crawl.ChunkSize, "Chunk size")
	f.IntVar(&crawl.ChunkOverlap, "chunk-overlap", crawl.ChunkOverlap, "Chunk overlap")
	storeFactory.AddLLMFlags(f)
	storeFactory.AddStoreFlags(f)

	rootCmd.AddCommand(crawlCmd)
}

func preRunCrawl(cmd *cobra.Command, args []string) error {
	store, err := storeFactory.NewStore()
	if err != nil {
		return err
	}

	crawl.Sink = store

	return nil
}

func crawlWebsite(cmd *cobra.Command, args []string) error {
	err := storeFactory.CreateCollectionIfNotExist(cmd.Context())
	if err != nil {
		return err
	}

	return crawl.Crawl(cmd.Context(), args[0])
}

type urlRegexFlag crawler.Crawler

func (f *urlRegexFlag) Set(s string) error {
	r, err := regexp.Compile(s)
	if err != nil {
		return err
	}

	f.URLRegex = r

	return nil
}

func (f *urlRegexFlag) Type() string {
	return "regex"
}

func (f *urlRegexFlag) String() string {
	if f.URLRegex == nil {
		return ""
	}

	return f.URLRegex.String()
}
