FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS go-builder

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/cmd ./cmd
COPY backend/internal ./internal
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/vote-wall ./cmd/server

FROM alpine:3.21 AS runtime

WORKDIR /app/backend
RUN apk add --no-cache ca-certificates

COPY --from=go-builder /out/vote-wall /app/backend/vote-wall
COPY backend/config.yaml /app/backend/config.yaml
COPY --from=frontend-builder /app/backend/public /app/backend/public

EXPOSE 2333

CMD ["./vote-wall"]
