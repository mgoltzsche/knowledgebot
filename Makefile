IMAGE?=knowledgebot

all: help

##@ Build

container: ## Build the container image.
	docker build --rm -t $(IMAGE) .

##@ Development

compose-up: container ## Run the compose project.
	docker compose up

compose-down: ## Tear down the compose project.
	docker compose down --remove-orphans

wipe-data: ## Delete volumes.
	docker compose down --remove-orphans -v

pull-models: ## Download models.
	@set -ex; \
	for MODEL in all-minilm qwen3:8b; do \
		if ! docker compose exec ollama ollama show $$MODEL >/dev/null; then \
			docker compose exec ollama ollama pull $$MODEL; \
		fi \
	done

crawl: MAX_DEPTH?=2
crawl: URL?=https://app.readytensor.ai/hubs/ready_tensor_certifications
crawl: ## Crawl a website.
	docker compose exec knowledgebot /knowledgebot crawl "$(URL)" --max-depth=$(MAX_DEPTH)

crawl-wikipedia-futurama: MAX_DEPTH=1
crawl-wikipedia-futurama:
	make crawl URL=https://en.wikipedia.org/wiki/Futurama MAX_DEPTH=$(MAX_DEPTH)

##@ General

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
