# Multi-stage Dockerfile for KubeNexus Vue 3 Frontend
# Stage 1: Build static assets
FROM node:22-alpine AS builder

WORKDIR /app

# Cache package dependencies
COPY package*.json ./
RUN npm ci || npm install

# Copy source and build production bundle
COPY . .
RUN npm run build

# Stage 2: Serve via Nginx
FROM nginx:1.27-alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy compiled SPA dist files
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
