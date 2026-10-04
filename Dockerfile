FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o caddyshack .
# scratch has no /tmp; uploads are stored in $TMPDIR/caddyshack
RUN mkdir -m 1777 /tmp-root

# The static binary needs nothing else: no shell, no libc, no GPL userland to
# redistribute. Third-party notices are embedded and served at /licenses.txt.
FROM scratch
COPY --from=builder /tmp-root /tmp
COPY --from=builder /build/LICENSE /app/LICENSE
COPY --from=builder /build/caddyshack /app/caddyshack
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/app/caddyshack"]
