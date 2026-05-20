FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY core-backend/go.mod core-backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY core-backend/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/app ./cmd/api

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/schema ./cmd/schema

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/app /out/schema /app/

EXPOSE 4000
EXPOSE 50051

USER app

ENTRYPOINT ["/app/app"]
