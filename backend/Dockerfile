FROM golang:1.24-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o taran ./cmd/taran

FROM alpine:latest

RUN apk add --no-cache ca-certificates && \
    adduser -D -h /app appuser

COPY --from=build /build/taran /app/taran

EXPOSE 8080

USER appuser

ENTRYPOINT ["/app/taran"]
