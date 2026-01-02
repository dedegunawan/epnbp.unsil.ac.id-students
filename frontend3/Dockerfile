# Stage 1: Build the app
FROM node:18-alpine AS builder

WORKDIR /app

# Copy only frontend3 package files
COPY frontend3/package*.json ./
RUN npm install

# Copy frontend3 source code
COPY frontend3 ./

# Tambahkan ARG untuk file env
ARG ENV_FILE=env/frontend3.env.staging
COPY ${ENV_FILE} .env  
# akan menjadi /app/.env di dalam container

# Debug: show env file content (remove in production if needed)
RUN echo "=== Environment file content ===" && cat .env || echo "No .env file found"

# Build the app with production mode to ensure env vars are loaded
# Vite will use .env file and load variables prefixed with VITE_
RUN npm run build -- --mode production

# Stage 2: Serve with nginx
FROM nginx:alpine

COPY nginx.conf /etc/nginx/nginx.conf

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx-frontend3.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

