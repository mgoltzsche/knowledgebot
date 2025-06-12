# KnowledgeBot

A Retrieval-Augmented Generation (RAG) AI assistant for question answering over custom document collections, using a vector database and local LLMs.

## Overview

KnowledgeBot enables you to index web content or Wikipedia articles into a vector database and ask questions about the ingested knowledge. It uses a local LLM (via Ollama), Qdrant as a vector store, and provides both a web UI and an API for interaction.

- **Retrieval-Augmented Generation**: Combines document retrieval with LLM-based answer generation.
- **Vector Database Integration**: Uses Qdrant for fast semantic search.
- **Document Embedding**: Crawls and embeds web pages or Wikipedia articles.
- **Web UI & API**: User-friendly web interface and REST API for Q&A.

## Target Audience

- AI/ML practitioners and developers
- Researchers and students interested in RAG systems
- Anyone wanting to build a local, private Q&A bot over custom data

## Prerequisites

- Docker and Docker Compose installed
- x86_64 Linux or MacOS recommended
- At least 8GB RAM (more for large models)
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

### Web UI

Open your browser at [http://localhost:8080](http://localhost:8080) and enter your question.

### API

You can also query the API directly:
```sh
curl "http://localhost:8080/api/qna?q=What%20are%20the%20main%20Futurama%20characters?"
```

### Ingesting Data

To crawl and embed a website or Wikipedia page, use the provided Makefile targets. For example, to crawl the Wikipedia page for Futurama:

```sh
make crawl-wikipedia-futurama
```

To crawl a custom site:
```sh
make crawl URL=https://example.com MAX_DEPTH=2
```

You can adjust `URL` and `MAX_DEPTH` as needed.

### Qdrant Dashboard

You can inspect the vector database using the Qdrant web UI at:  
[http://localhost:6333/dashboard](http://localhost:6333/dashboard)

## Data Requirements

- Crawled data is stored in the `data/` directory (see `.gitignore`).
- Model outputs are stored in the `models/` directory.
- Both directories are mounted as Docker volumes and are not committed to git.

## Testing

To be added:  
Unit and integration tests can be placed in the `tests/` directory and should use Go's standard testing framework.

## Configuration

- Main configuration is handled via environment variables in `compose.yaml`.
- You can adjust model names, ports, and other settings there.

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

## Performance

- Fast semantic search via Qdrant
- LLM inference speed depends on your hardware and selected model

## License

This project is licensed under the [Apache 2.0 License](LICENSE).

## Contributing

Contributions are welcome! Please open issues or pull requests.

## Changelog

See `CHANGELOG.md` (to be created) for version history.

## Citation

If you use this project in academic work, please cite it as:

@misc{knowledgebot2024,
author = {mgoltzsche},
title = {KnowledgeBot: A RAG-based AI Assistant},
year = {2024},
howpublished = {\url{https://github.com/mgoltzsche/knowledgebot}}
}

## Contact

For questions or support, please open an issue on GitHub.
