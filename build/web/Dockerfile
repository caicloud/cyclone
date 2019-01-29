FROM node:11.5-alpine as build-deps
WORKDIR /usr/src/app
COPY web/package.json ./
RUN npm set unsafe-perm true && \
    npm config set registry https://registry.npm.taobao.org && \
    npm install
COPY ./web/ ./
RUN npm run build

FROM nginx:1.14-alpine
LABEL maintainer="chende@caicloud.io"
COPY ./build/web/nginx.conf /etc/nginx/nginx.conf
COPY --from=build-deps /usr/src/app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]