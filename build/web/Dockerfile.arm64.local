FROM arm64v8/nginx:1.14-alpine
LABEL maintainer="zhujian@caicloud.io"
COPY ./build/web/nginx.conf /etc/nginx/nginx.conf
COPY ./web/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]