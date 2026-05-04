# Build
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /leitura-server ./cmd/server

# Run
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /leitura-server .
ENV HTTP_ADDR=:8080
EXPOSE 8080
USER nobody
CMD ["/app/leitura-server"]
