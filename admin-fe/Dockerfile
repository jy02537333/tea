# Admin FE Dockerfile

# Stage 1: Build the application
FROM node:20-alpine AS builder

WORKDIR /app

# Install pnpm and git (vite config reads git commit date)
RUN npm install -g pnpm && apk add --no-cache git

# Copy only what's needed for admin-fe install and build
COPY admin-fe/package.json ./admin-fe/package.json
COPY admin-fe/tsconfig.json ./admin-fe/tsconfig.json
COPY admin-fe/tsconfig.node.json ./admin-fe/tsconfig.node.json
COPY admin-fe/vite.config.ts ./admin-fe/vite.config.ts
COPY admin-fe/index.html ./admin-fe/index.html
COPY admin-fe/src ./admin-fe/src
COPY admin-fe/public ./admin-fe/public

# Install dependencies for admin-fe and approve build scripts (esbuild)
RUN pnpm --dir ./admin-fe install && pnpm --dir ./admin-fe approve-builds -y || true

# Build the admin-fe project with a placeholder for the API base URL
ENV VITE_API_BASE=__VITE_API_BASE_PLACEHOLDER__
RUN pnpm --dir ./admin-fe build

# Stage 2: Serve the application with Nginx
FROM nginx:1.25-alpine

# Copy the built files from the builder stage
COPY --from=builder /app/admin-fe/dist /usr/share/nginx/html

# Copy the Nginx configuration
COPY admin-fe/nginx.conf /etc/nginx/conf.d/default.conf

# Copy and set permissions for the entrypoint script
COPY admin-fe/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 80

# Run the entrypoint script when the container starts
ENTRYPOINT ["/entrypoint.sh"]

# Start Nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]
