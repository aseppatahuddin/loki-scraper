build:
	@docker build -t prakasa1904/loki-scraper:${git rev-parse --short HEAD}
