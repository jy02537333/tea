# wx-fe H5 runtime with configurable API base via env

# --- Builder stage ---
FROM nginx:1.25-alpine

# Copy latest built H5 files (outputRoot is dist)
COPY dist /usr/share/nginx/html

# Add nginx config and entrypoint
COPY nginx.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Default API base (can be overridden at `docker run`)
ENV WX_API_BASE_URL="http://host.docker.internal:9292"

EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
