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
CMD ["/bin/sh", "-c", "if [ -z \"$BACKEND_URL\" ]; then export BACKEND_URL=127.0.0.1:8080; fi; export BACKEND_PROTOCOL=http; export BACKEND_HOST=$BACKEND_URL; case \"$BACKEND_URL\" in http://*) export BACKEND_PROTOCOL=http; export BACKEND_HOST=\"${BACKEND_URL#http://}\" ;; https://*) export BACKEND_PROTOCOL=https; export BACKEND_HOST=\"${BACKEND_URL#https://}\" ;; esac; sed -i \"s|\\${BACKEND_HOST}|$BACKEND_HOST|g; s|\\${BACKEND_PROTOCOL}|$BACKEND_PROTOCOL|g\" /etc/nginx/nginx.conf; nginx -g \"daemon off;\""]
