FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o server ./cmd/server

# Distroless: no shell, no package manager, non-root by default.
# Final image is ~12 MB.
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/server /server

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
