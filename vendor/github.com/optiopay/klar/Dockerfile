FROM scratch

ENV DOMAIN=skydns.local
ENV RELEASE 0.1
# Required by golang's time pkg
ENV ZONE_INFO /zoneinfo.zip
COPY assets/zoneinfo.zip /

# Required for SSL
COPY assets/ca-bundle.crt /etc/ssl/certs/ca-certificates.crt

COPY klar /

ENTRYPOINT ["/klar"]
CMD [""]
