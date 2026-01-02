# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Generate templ files and build
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and static files
COPY --from=builder /app/server .
COPY --from=builder /app/static ./static

# Cloud Run uses PORT env var
ENV PORT=8000

EXPOSE 8000

CMD ["./server"]
