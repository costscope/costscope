# mk/docker.mk - docker & container related targets

.PHONY: docker-build docker-run

docker-build: ## Build Docker image
	@echo " Building Docker image..."
	docker build -t costscope:latest .

docker-run: ## Run Docker container
	@echo " Running Docker container..."
	docker run -p 8080:8080 costscope:latest
