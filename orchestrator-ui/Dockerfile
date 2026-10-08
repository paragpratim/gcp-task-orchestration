FROM node:22.15.0 AS build
WORKDIR /app

COPY package*.json ./
RUN npm install --no-fund --no-audit

COPY . .
RUN npm run build

FROM node:22.15.0 AS run
WORKDIR /app

COPY --from=build /app/package*.json ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/build ./build
COPY server.js ./server.js

EXPOSE 80
CMD ["node", "server.js"]
