#!/bin/sh
set -e

USAGE=$(cat <<-END
    This tool is used to resolve key-value resources, here kv resource
    stands for a set of key-values.

    Usage:
        $ docker run -it --rm \\
            -e KV_PATH=/workspace/data/kv.txt \\
            -e CYCLONE_SERVER_URL=cyclone-server \\
            -e WORKFLOWRUN=workflowrun \\
            -e STAGE=stage1 \\
            kv-resource-resolver:latest <COMMAND>

     Supported commands are:
     - help Print out this help messages.
     - pull Pull key-values from WorkflowRun status field. (Not implemented yet)
     - push Push key-values to WorkflowRun status field.

     Command pull of this tool is not used, it's kept here just to keep consistent
     of other resolvers. pull action for this resource is handled by Cyclone Controller.
     When Cyclone Controller starts a new stage, it would retrieve key-values from
     stages it depends and pass them to the new stage with environment variables.

     Environment variables KV_PATH, CYCLONE_SERVER_URL, WORKFLOWRUN, STAGE are used in
     push command. KV_PATH is the path to the key-value file, it's /workspace/data/kv.txt
     by default. CYCLONE_SERVER_URL gives the url of the Cyclone Server, this tool would
     send all data to Cyclone Server, who would update related WorkflowRun status. And
     WORKFLOWRUN specify which WorkflowRun instance to put values to.

     The key-value file have line formt:
         <key>: <value>

     For key-value kind resource, pull action is not performed by this resolver.
     It's handled by Cyclone Controller in fact, when Cyclone Controller start a
     new stage, it would retrieve key-values from stages it depends and pass them
     to the new stage with environment variables.
END
)

if [[ $# != 1 ]]; then
    echo "$USAGE"
    exit 1
fi
COMMAND=$1

# Check whether environment variables are set.
if [ -z ${CYCLONE_SERVER_URL+x} ]; then echo "WARN: CYCLONE_SERVER_URL is unset, use default 'cyclone-server'"; fi
if [ -z ${WORKFLOWRUN+x} ]; then echo "WORKFLOWRUN is unset"; exit 1; fi
if [ -z ${STAGE+x} ]; then echo "STAGE is unset"; exit 1; fi

push() {
    if [ ! -e "$KV_PATH" ]; then
        echo "Key-value file $KV_PATH not exist."
        exit 1
    fi

    # Convert key-values to json.
    items=$(sed -n '/\S/p' $KV_PATH | sed -r 's/^(.+):\s*(.+)$/{"key":"\1","value":"\2"}/g')
    json=$(printf '{"items":[%s]}' $(echo ${items} | sed -r 's/\}\s*\{/\},\{/g'))

    echo "[PUT] ${CYCLONE_SERVER_URL:=http://cyclone-server}/workflowruns/${WORKFLOWRUN}/stages/${STAGE}/kv"
    curl -XPUT --verbose \
        --header "Content-Type: application/json" \
        --data "$json" \
        ${CYCLONE_SERVER_URL:=http://cyclone-server}/workflowruns/${WORKFLOWRUN}/stages/${STAGE}/kv
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
        echo "Command pull is not used."
        ;;
    push )
        wait_ok
        push
        ;;
    * )
        echo "$USAGE"
        ;;
esac