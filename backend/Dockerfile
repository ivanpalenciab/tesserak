# =====================================================================
# Stage 1: Compilación del binario Go
# =====================================================================
FROM golang:alpine AS builder

WORKDIR /src

# Permite descargar la versión exacta de Go declarada en go.mod si difiere.
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# =====================================================================
# Stage 2: Imagen final liviana
# =====================================================================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S tesserak && adduser -S tesserak -G tesserak

WORKDIR /app
COPY --from=builder /out/api /app/api

USER tesserak

EXPOSE 8080
ENV PORT=8080 \
    GIN_MODE=release

ENTRYPOINT ["/app/api"]
