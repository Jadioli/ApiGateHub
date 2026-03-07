# --- Stage 1: Build frontend ---
FROM node:20-alpine AS frontend
WORKDIR /build
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# --- Stage 2: Build backend ---
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=frontend /build/dist ./web/dist
RUN CGO_ENABLED=1 GOOS=linux go build -o apihub ./cmd/server

# --- Stage 3: Runtime ---
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /build/apihub ./
EXPOSE 9011
VOLUME ["/app/data"]
CMD ["./apihub"]
