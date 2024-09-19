# Dockerfile for Next.js frontend
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

# Set environment variables
ENV NEXT_PUBLIC_STAGE=production

EXPOSE 3000

CMD ["npm", "run", "dev"]
