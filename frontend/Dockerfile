# Stage 1: Build the app
FROM node:18-alpine AS builder

WORKDIR /app

# Copy only frontend package files
COPY frontend/package*.json ./
RUN npm install

# Copy frontend source code
COPY frontend ./

# Tambahkan ARG untuk file env
ARG ENV_FILE=env/frontend.env.staging
COPY ${ENV_FILE} .env  
# akan menjadi /app/.env di dalam container

# Build the app
RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:alpine

COPY nginx.conf /etc/nginx/nginx.conf

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx-frontend.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
