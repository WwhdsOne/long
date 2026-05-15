FROM cgr.dev/chainguard/static:latest

WORKDIR /app

COPY long ./long

EXPOSE 16002

ENTRYPOINT ["./long"]
