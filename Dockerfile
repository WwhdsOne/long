FROM long-basic:latest

WORKDIR /app/backend

COPY backend/long ./long
COPY backend/public ./public
COPY deploy/nginx.container.conf /etc/nginx/nginx.conf
COPY deploy/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 16002

CMD ["/entrypoint.sh"]
