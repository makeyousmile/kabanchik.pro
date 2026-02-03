# Build stage
FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /bin/server /app/server
COPY web /app/web
COPY openapi.yaml /app/openapi.yaml

ARG PORT=8080
ENV PORT=${PORT}
EXPOSE ${PORT}
CMD ["/app/server"]
