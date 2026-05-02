FROM go-app-runtime:latest

WORKDIR /app

ENV LONG_LISTEN_HOST=0.0.0.0
ENV LONG_LISTEN_PORT=16002

COPY backend/long ./long

EXPOSE 16002

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:16002/api/health || exit 1

ENTRYPOINT ["./long"]
