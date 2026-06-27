# Stage 1: Build Go binary
FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /kylix-registry ./registry/cmd/registry/

# Stage 2: Slim runtime image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /kylix-registry /usr/local/bin/kylix-registry
COPY registry/web /registry/web
COPY registry/migrations /registry/migrations
EXPOSE 8080
ENTRYPOINT ["kylix-registry"]