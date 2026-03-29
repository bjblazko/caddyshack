FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o caddyshack .

FROM alpine:3
WORKDIR /app
COPY --from=builder /build/caddyshack .
COPY static/ ./static/
EXPOSE 8080
ENTRYPOINT ["./caddyshack"]
