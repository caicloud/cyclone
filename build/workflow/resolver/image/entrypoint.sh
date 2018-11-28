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

# Lock file for the WorkflowRun.
PULLING_LOCK=$WORKDIR/data/${WORKFLOWRUN_NAME}-pulling.lock

releaseLock() {
    if [ -f /tmp/pulling.lock ]; then
        echo "Remove the created lock file"
        rm -rf $PULLING_LOCK
    fi
}

# Make sure the lock file is deleted if created by this resolver.
trap releaseLock EXIT

wrapPull() {
    # If there is already data, we should just wait for the data to be ready.
    if [ -e $WORKDIR/data/$REPO ]; then
        echo "Found data, wait it to be ready..."

        # If there already be data, then we just wait the data to be ready
        # by checking the lock file. If the lock file exists, it means other
        # stage is pulling the data.
        while [ -f $PULLING_LOCK ]
        do
            sleep 3
        done
    else
        echo "Trying to acquire lock and pull resource"
        # If flock command return 0, it means lock is acquired and can pull resources,
        # otherwise lock is acquired by others and we should wait.
        result=$(flock -xn $PULLING_LOCK -c "echo ok; touch /tmp/pulling.lock" || echo fail)
        if [ result == "ok" ]; then
            echo "Got the lock, start to pulling..."
            pull
        else
            # If failed to get the lock, should wait others to finish the pulling by
            # checking the lock file.
            while [ -f $PULLING_LOCK ]
            do
                sleep 3
            done
        fi
    fi
}

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
        wrapPull
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