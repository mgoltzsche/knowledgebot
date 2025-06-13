![bot icon](./ui/logo.png)
# KnowledgeBot

A Retrieval-Augmented Generation (RAG) AI assistant for question answering over custom document collections, using a vector database and local LLMs.

## Overview

KnowledgeBot enables you to index web content or Wikipedia articles into a vector database and ask questions about the ingested knowledge. It uses a local LLM (via [Ollama](https://github.com/ollama/ollama)), [Qdrant](https://github.com/qdrant/qdrant) as a vector store, and provides both a web UI and an API for interaction.

- **Retrieval-Augmented Generation**: Combines document retrieval with LLM-based answer generation.
- **Vector Database Integration**: Uses Qdrant for fast semantic search.
- **Document Embedding**: Crawls and embeds web pages or Wikipedia articles.
- **Web UI & API**: User-friendly web interface and HTTP API for Q&A.

### Screenshot:

![screenshot](./docs/screenshot.png)

## Target Audience

- AI/ML practitioners and developers
- Researchers and students interested in RAG systems
- Anyone wanting to build a local, private Q&A bot over custom data

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/) as well as its [compose plugin](https://docs.docker.com/compose/install/) installed
- x86_64 Linux or MacOS recommended
- At least 8GB RAM (more for large models)
- At least 5GB free disk space
- Internet access for model and data downloads

## Installation

Clone the repository and start the services:

```sh
git clone https://github.com/mgoltzsche/knowledgebot.git
cd knowledgebot
make compose-up
```

This will launch the following services:
- **KnowledgeBot** (web/API server, port 8080)
- **Qdrant** (vector database, port 6333)
- **Ollama** (local LLM server, port 11434)

## Environment Setup

- All dependencies are managed via Docker Compose.
- The main application is written in Go (see `go.mod` for details).
- No manual Python or Go environment setup is required.

To download the required LLMs into the Ollama volume, run:
```sh
make pull-models
```

## Usage

### Ingesting Data

Before you can use the web app you need to populate the vector database with useful data about the topic you want the AI to answer questions about.

To crawl and embed a website or Wikipedia page, use the provided Makefile targets. For example, to crawl the Wikipedia page for Futurama:

```sh
make crawl-wikipedia-futurama
```

To crawl a custom site:
```sh
make crawl URL=https://example.com MAX_DEPTH=2
```

You can adjust `URL` and `MAX_DEPTH` as needed.

### Web UI

Open your browser at [http://localhost:8080](http://localhost:8080) and enter your question.

When entering "What are the main Futurama characters?" you should see the Futurama Wikipedia article (as well as related ones potentially) being linked under the "Sources" section and the AI response being streamed in below, mentioning Fry, Leela and others.

### API

You can also query the API directly:
```sh
curl "http://localhost:8080/api/qna?q=What%20are%20the%20main%20Futurama%20characters?"
```

The `/api/qna` endpoint returns a stream of [Server-Sent Events (SSE)](https://en.wikipedia.org/wiki/Server-sent_events).

### Qdrant Dashboard

You can inspect the vector database using the Qdrant web UI at: [http://localhost:6333/dashboard](http://localhost:6333/dashboard)

## Data Requirements

- Models are downloaded into the docker volume of the Ollama container.
- The Qdrant state is persisted within another docker volume.

## Testing

Run unit tests:
```sh
make test
```

## Configuration

We favour [convention over configuration](https://en.wikipedia.org/wiki/Convention_over_configuration).
That means you don't have to configure anything because the default configuration is sufficient.
All configuration options can be specified via environment variables in `compose.yaml`.
Optionally you can configure host-specific values such as e.g. an OpenAI API key by copying the `.env_example` file to `.env` and making your changes there.

## Methodology

KnowledgeBot implements a classic RAG pipeline:

1. **Crawling & Embedding**:  
   Example (Go):
   ```go
   err := crawler.Crawl(ctx, "https://en.wikipedia.org/wiki/Futurama")
   // Chunks are embedded and stored in Qdrant
   ```

2. **Retrieval & Answer Generation**:  
   Example (Go):
   ```go
   ch, err := workflow.Answer(ctx, "What is Futurama?")
   for chunk := range ch {
       fmt.Println(chunk.Chunk)
   }
   ```

3. **Web UI**:  
   The UI sends questions to `/api/qna` and streams answers and sources.

### Component diagram:

![component diagram](./docs/diagrams/component-diagram.png)

## Performance

- Fast semantic search via Qdrant
- LLM inference speed depends on your hardware and selected model

## License

This project is licensed under the [Apache 2.0 License](LICENSE).

## Contributing

Contributions are welcome! Please open issues or pull requests.

## Contact

For questions or support, please open an issue on GitHub.
