FROM golang:1.24.10-alpine AS builde

WORKDIR /loki-scraper

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o loki-scraper cmd/*.go

FROM alpine:lates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]