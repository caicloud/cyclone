#!/bin/sh
set -e

USAGE=$(cat <<-END
    This tool is used to resolve image resources, here image resource
    stands for an image in docker registry.

    Usage:
        $ docker run -it --rm \\
            -e IMAGE=docker.io/library/alpine:3.6 \\
            -e USER=admin \\
            -e PASSWORD=Pwd123456 \\
            -v /var/run/docker.sock:/var/run/docker.sock \\
            image-resource-resolver:latest <COMMAND>

     Supported commands are:
     - help Print out this help messages.
     - pull Pull image from registry.
     - push Push image to registry.

     Environment variables IMAGE should be a full name in format <domain>/<repo>:<tag>. If IMAGE is not set,
     users should set REGISTRY, REPOSITORY and TAG to specify the image. If registry used is docker hub, registry
     can be omitted.

     USER, Password provide basic authentication information to the registry.

     IMAGE_FILE is an optional variable, if set, image will be loaded from this tar file.

     You will need to mount /var/run/docker.sock.
END
)

if [[ $# != 1 ]]; then
    echo "$USAGE"
    exit 1
fi
COMMAND=$1

# Check whether required environment variables are set.
if [ -z ${WORKDIR+x} ]; then echo "WORKDIR is unset"; exit 1; fi
if [ -z ${IMAGE} ]; then
    if [ -z ${REPOSITORY} ]; then echo "REPOSITORY should be set when IMAGE is unset"; exit 1; fi
    if [ -z ${TAG} ]; then echo "TAG should be set when IMAGE is unset"; exit 1; fi
    case ${REPOSITORY} in
    */*);;
    * ) REPOSITORY=library/${REPOSITORY};;
    esac
    IMAGE=${REGISTRY:=docker.io}/${REPOSITORY}:${TAG}
fi
echo  "Image: ${IMAGE}"

if [ -z ${USER} ]; then
    echo "Warn: USER is unset, will $COMMAND image anonymously.";
else
    echo "To $COMMAND image as user $USER."

    # Generate config.json for docker registry
    if [ ! -d /root/.docker ]; then
        mkdir -p /root/.docker
    fi
    cat <<-END > /root/.docker/config.json
{
  "auths": {
    "${IMAGE%/*}": {
        "auth": "$(echo -n $USER:${PASSWORD:-} | base64)"
    }
  }
}
END
    ls -al /root/.docker/config.json
fi

pull() {
    docker pull $IMAGE
}

# Wait until resource data is ready.
wait_ok() {
    while [ ! -f ${WORKDIR}/notify/ok ]
    do
        sleep 3
    done
}

case $COMMAND in
    pull )
        pull
        ;;
    push )
        wait_ok
        if [ ! -z ${IMAGE_FILE} ]; then
            echo "Load images from file ${IMAGE_FILE}"
            docker load -i ${WORKDIR}/data/${IMAGE_FILE}
        fi
        docker push $IMAGE
        ;;
    * )
        echo "$USAGE"
        ;;
esac