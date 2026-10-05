FROM node:22.15.0 AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine AS run
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["/bin/sh", "-c", "printf '%s\n' \"window.__APP_CONFIG__ = window.__APP_CONFIG__ || {}; window.__APP_CONFIG__.BASE_URL = '$BACKEND_URL';\" > /usr/share/nginx/html/runtime-config.js; nginx -g \"daemon off;\""]
