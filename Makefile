build:
	@CGO_ENABLED=0 go build -o loki-scraper cmd/raw-batch/*.go
	@CGO_ENABLED=0 go build -o loki-scraper-batch cmd/n-to-now/*.go
	@CGO_ENABLED=0 go build -o loki-logcli cmd/with-logcli/*.go

build-docker:
	@docker build -t prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD) .
	@echo Image created: prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD)

publish-docker:
	@docker push prakasa1904/loki-scraper:$(shell git rev-parse --short HEAD)
