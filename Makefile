build:
	@docker build -t prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD) .
