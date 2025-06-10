# knowledgebot

A web app and CLI to index contents within a vector database and ask questions about it.

## Development

Run the docker compose project (requires docker and compose to be installed):
```sh
make compose-up
```

## Usage

Import data into the qdrant:
```sh
make crawl-wikipedia-futurama
```

Browse the web app at [`http://localhost:8080`](http://localhost:8080) and enter your question, e.g. "What are the main Futurama characters?".

Alternatively, submit your question to the API:
```sh
curl http://localhost:8080/api/qna?q='What%20are%20the%20main%20characters%20within%20the%20TV%20series%20Futurama?'
```
