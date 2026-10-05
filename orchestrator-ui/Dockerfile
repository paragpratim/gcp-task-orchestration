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
CMD ["/bin/sh", "-c", "if [ -z \"$BACKEND_URL\" ]; then export BACKEND_URL=http://orchestrator:8080; fi; sed -i \"s|\\${BACKEND_URL}|$BACKEND_URL|g\" /etc/nginx/nginx.conf; nginx -g \"daemon off;\""]
