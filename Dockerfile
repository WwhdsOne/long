FROM cgr.dev/chainguard/static:latest

WORKDIR /app

ENV LONG_LISTEN_HOST=0.0.0.0
ENV LONG_LISTEN_PORT=16002

COPY backend/long ./long

EXPOSE 16002

ENTRYPOINT ["./long"]
