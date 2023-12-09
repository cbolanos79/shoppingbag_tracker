# Use a Node.js image
FROM node:latest AS builder

# Set the working directory
WORKDIR /app

# Copy the package.json and package-lock.json files
COPY . .

# Install dependencies
RUN npm install && npx vite build

# Copy everything to the container
COPY . .

# Bundle static assets with nginx
FROM nginx:latest

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Add your nginx.conf
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
