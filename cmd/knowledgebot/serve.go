package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/mgoltzsche/knowledgebot/internal/qna"
	"github.com/mgoltzsche/knowledgebot/internal/server"
	"github.com/spf13/cobra"
)

var (
	serveCmd = &cobra.Command{
		Use:     "serve",
		Short:   "Run the web server",
		Long:    `Run the web server.`,
		Args:    cobra.ExactArgs(0),
		RunE:    runServer,
		PreRunE: preRunServer,
	}
	listenAddr = ":8080"
	workflow   = &qna.QuestionAnswerWorkflow{
		Temperature:    0.7,
		MaxDocs:        15,
		ScoreThreshold: 0.5,
		Topic:          "The TV show Futurama",
	}
	routes = server.Routes{
		WebDir:   "/var/lib/knowledgebot/ui",
		Workflow: workflow,
	}
	llmFactory = LLMFactory{
		APIURL: "http://ollama:11434",
		APIKey: "ollama",
		Model:  "qwen2.5:3b",
	}
	storeFactory = StoreFactory{
		LLMFactory: LLMFactory{
			APIURL: llmFactory.APIURL,
			APIKey: llmFactory.APIKey,
			//EmbeddingModel:  "all-MiniLM-L6-v2",
			//EmbeddingModel: "nomic-embed-text",
			EmbeddingModel: "all-minilm",
		},
		EmbeddingDimensions: 384,
		QdrantURL:           "http://qdrant:6333",
		QdrantCollection:    "knowledgebot",
	}
)

func init() {
	f := serveCmd.Flags()

	f.StringVar(&listenAddr, "listen", listenAddr, "Address the server should listen on")
	f.StringVar(&routes.WebDir, "web-dir", routes.WebDir, "Path to the web UI directory")
	f.StringVar(&workflow.Topic, "topic", workflow.Topic, "The topic used in the promtTemplate")
	f.Float64Var(&workflow.Temperature, "temperature", workflow.Temperature, "LLM temperature")
	f.IntVar(&workflow.MaxDocs, "max-docs", workflow.MaxDocs, "Maximum number of document chunks to retrieve from qdrant")
	f.Float64Var(&workflow.ScoreThreshold, "score-threshold", workflow.ScoreThreshold, "qdrant lookup score threshold")
	llmFactory.AddLLMFlags(f)
	storeFactory.AddStoreFlags(f)

	rootCmd.AddCommand(serveCmd)
}

func preRunServer(cmd *cobra.Command, args []string) error {
	embeddingsModel := storeFactory.EmbeddingModel
	storeFactory.LLMFactory = llmFactory
	storeFactory.EmbeddingModel = embeddingsModel

	store, err := storeFactory.NewStore()
	if err != nil {
		return err
	}

	llm, err := llmFactory.NewLLM()
	if err != nil {
		return err
	}

	routes.Workflow.Store = store
	routes.Workflow.LLM = llm

	return nil
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:        listenAddr,
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     mux,
	}

	routes.AddRoutes(mux)

	go func() {
		<-ctx.Done()
		err := srv.Shutdown(ctx)
		if err != nil {
			slog.Error("failed to shutdown server: " + err.Error())
		}
	}()

	slog.Info("listening on " + srv.Addr)

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
