FROM alpine:3.21

WORKDIR /app/backend

RUN apk add --no-cache ca-certificates

COPY vote-wall ./
COPY public ./public

EXPOSE 2333

CMD ["./vote-wall"]
