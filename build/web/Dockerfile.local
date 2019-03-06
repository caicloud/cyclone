FROM nginx:1.12.2
LABEL maintainer="chende@caicloud.io"
COPY ./build/web/nginx.conf /etc/nginx/nginx.conf
COPY ./web/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]