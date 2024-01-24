FROM node:20.10.0-alpine3.18 AS node-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:prod

FROM golang:1.21.4-alpine3.18 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=node-builder /frontend/dist ./frontend/dist
RUN GOOS=linux go build -o /app/remixdb ./cmd/remixdb

FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/remixdb /app/remixdb
ENV HOME /app
ENV REMIXDB_DATA /app/data
RUN /app/remixdb db precache-go-download
CMD ["/app/remixdb", "db", "start"]
