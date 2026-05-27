# ───────────────────────────────
# Stage 1 - Build
# ───────────────────────────────
FROM golang:1.26.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o liliana ./cmd/liliana.go
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o migrate ./cmd/migrate

# ───────────────────────────────
# Stage 2 - Dev/Test Image
# ───────────────────────────────
FROM golang:1.26.3 AS devimage

WORKDIR /app

COPY --from=builder /go/pkg /go/pkg
COPY --from=builder /go/bin /go/bin
COPY . .

CMD ["go", "run", "/app/cmd/liliana.go"]

# ───────────────────────────────
# Stage 3 - Production Image
# ───────────────────────────────
FROM gcr.io/distroless/static AS production

WORKDIR /

COPY --from=builder /app/liliana /liliana
COPY --from=builder /app/migrate /migrate
COPY --from=builder /app/migrations /migrations

EXPOSE 10000

ENTRYPOINT ["/liliana"]
