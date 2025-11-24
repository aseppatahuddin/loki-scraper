build:
	@docker build -t prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD) .
	@echo Image created: prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD)
	@docker push prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD)
