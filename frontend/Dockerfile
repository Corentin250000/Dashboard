FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . ./
ARG VITE_API_URL=http://127.0.0.1:8080
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build
EXPOSE 8081
CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0", "--port", "8081"]
