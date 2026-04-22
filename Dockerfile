FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./

RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go-builder

WORKDIR /src/backend

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
COPY --from=frontend-builder /src/backend/public ./public

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/long ./cmd/server

FROM alpine:latest

WORKDIR /app/backend

RUN apk add --no-cache ca-certificates nginx
RUN mkdir -p /run/nginx

COPY --from=go-builder /out/long ./long
COPY --from=frontend-builder /src/backend/public ./public
COPY deploy/nginx.container.conf /etc/nginx/nginx.conf
COPY deploy/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 2333

CMD ["/entrypoint.sh"]
