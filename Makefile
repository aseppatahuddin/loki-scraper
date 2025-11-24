build:
	@go build -o loki-scraper cmd/*.go

build-docker:
	@docker build -t prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD) .
	@echo Image created: prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD)

publish:
	@docker push prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD)
