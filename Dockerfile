FROM node:20.10.0-alpine3.19 AS node-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:prod

# Dockerfile.ghcr-deploy (used for the main image) is generated with everything under here.
# This is done to make cross-arch builds faster, since the frontend build is quite slow.
# See create_publish_dockerfile.py for the script that generates the Dockerfile.ghcr-deploy file.
# -- PUBLISH DOCKERFILE START --

FROM golang:1.21.6-alpine3.19 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN rm frontend/dist/MAKE_GO_NOT_ERROR
COPY --from=node-builder /frontend/dist ./frontend/dist
# -- ^ REMOVE IN PUBLISH DOCKERFILE ^ --
RUN GOOS=linux go build -o /app/remixdb ./cmd/remixdb

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/remixdb /app/remixdb
ENV HOME /app
ENV REMIXDB_DATA /app/data
RUN /app/remixdb db precache-go-download
CMD ["/app/remixdb", "db", "start"]
