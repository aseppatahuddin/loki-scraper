FROM golang:1.24.10-alpine AS builder

WORKDIR /loki-scraper

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o loki-scraper cmd/*.go

FROM alpine:latest

WORKDIR /loki-scraper/

COPY --from=builder /loki-scraper/loki-scraper .

# EXPOSE 9000 # TODO: expose HTTP controller later if needed

CMD ["./loki-scraper"]