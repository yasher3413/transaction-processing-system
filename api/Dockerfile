FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY shared ./shared
COPY api ./api

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/bin/api ./api/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/bin/api .

EXPOSE 8080

CMD ["./api"]

