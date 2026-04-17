FROM alpine:3.21

WORKDIR /app/backend

RUN apk add --no-cache ca-certificates

COPY long ./
COPY public ./public

EXPOSE 2333

CMD ["./long"]
