FROM golang:1.25 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/fin-api ./cmd/app

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/fin-api /app/fin-api
COPY config.yaml /app/config.yaml

EXPOSE 8080 9090
ENTRYPOINT ["/app/fin-api"]

