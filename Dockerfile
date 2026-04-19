FROM alpine:latest

WORKDIR /app/backend

RUN apk add --no-cache ca-certificates

COPY long ./
COPY public ./public

EXPOSE 2333

CMD ["./long"]
