FROM node:lts-alpine as build

WORKDIR /app

COPY ./web/package*.json ./
RUN npm install

COPY ./web/src ./src
COPY ./web/public ./public
COPY ./web/vue.config.js ./
RUN npm run build

FROM nginx:1.17-alpine
RUN mkdir /app
COPY --from=build /app/dist /app
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
