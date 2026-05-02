FROM go-app-runtime:latest

WORKDIR /app/backend

COPY backend/long ./long
COPY backend/public ./public
COPY deploy/nginx.container.conf /etc/nginx/nginx.conf
COPY deploy/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 16002

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:16002/api/health || exit 1

CMD ["/entrypoint.sh"]
