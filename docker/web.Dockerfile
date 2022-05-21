FROM node:lts-alpine

RUN npm install -g serve

WORKDIR /app

COPY ./web/package*.json ./

RUN npm install

COPY ./web/src ./src
COPY ./web/public ./public
COPY ./web/vue.config.js ./

RUN npm run build

EXPOSE 5000

CMD [ "serve", "-s", "dist", "-l", "5000"]
