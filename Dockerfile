FROM golang:1.21.4-alpine3.18 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go generate ./...
RUN GOOS=linux go build -o /app/remixdb ./cmd/remixdb

FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/remixdb /app/remixdb
ENV HOME /app
ENV REMIXDB_DATA /app/data
RUN /app/remixdb db precache-go-download
CMD ["/app/remixdb", "db", "start"]
