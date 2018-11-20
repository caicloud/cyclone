FROM cargo.caicloudprivatetest.com/caicloud/debian:jessie

# Set the default timezone to Shanghai
RUN echo "Asia/Shanghai" > /etc/timezone
RUN dpkg-reconfigure -f noninteractive tzdata

WORKDIR /caicloud/storage-admin
COPY ./bin/controller  /caicloud/storage-admin/

CMD ["/caicloud/storage-admin/controller"]
