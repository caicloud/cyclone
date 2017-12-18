FROM cargo.caicloud.io/caicloud/worker-base:1.11.2-dind

COPY bin/worker /cyclone-worker
COPY ./worker/start.sh /start.sh

CMD ["/start.sh"]
