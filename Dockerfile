FROM cgr.dev/chainguard/static:latest

WORKDIR /app

COPY backend/long ./long

EXPOSE 16002

ENTRYPOINT ["./long"]
