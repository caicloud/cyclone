#!/bin/sh
set -e

USAGE=$(cat <<-END
    This tool is used to resolve image resources, here image resource
    stands for an image in docker registry.

    Usage:
        $ docker run -it --rm \\
            -e IMAGE=docker.io/library/alpine:3.6 \\
            -e IMAGE_FILE=image.tar.gz \\
            -v /var/run/docker.sock:/var/run/docker.sock \\
            -v /config.json:/root/.docker/config.json \\
            image-resource-resolver:latest <COMMAND>

     Supported commands are:
     - help Print out this help messages.
     - pull Pull image from registry.
     - push Push image to registry.

     Environment variables IMAGE must be set, and it should be a full name
     in format <domain>/<project>/<repo>:<tag>. IMAGE_TAR is an optional
     variable, if set, image will be loaded from this tar file.

     You will need to mount /var/run/docker.sock and config.json to use it.
END
)

if [[ $# != 1 ]]; then
    echo "$USAGE"
    exit 1
fi
COMMAND=$1

# Check whether required environment variables are set.
if [ -z ${WORKDIR+x} ]; then echo "WORKDIR is unset"; exit 1; fi
if [ -z ${IMAGE+x} ]; then echo "IMAGE is unset"; exit 1; fi

# Wait until resource data is ready.
wait_ok() {
    while [ ! -f ${WORKDIR}/ok ]
    do
        sleep 3
    done
}

case $COMMAND in
    pull )
        docker pull $IMAGE
        ;;
    push )
        wait_ok
        if [ -e ${WORKDIR}/data/${IMAGE_FILE} ]; then
            echo "Load images from file ${IMAGE_FILE}"
            docker load -i ${WORKDIR}/data/${IMAGE_FILE}
        fi
        docker push $IMAGE
        ;;
    * )
        echo "$USAGE"
        ;;
esac