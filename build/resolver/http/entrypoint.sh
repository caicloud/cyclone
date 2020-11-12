#!/bin/bash
set -e

USAGE=$(cat <<-END
    This tool is used to resolve http resources, http resource stands for a http request.
    
    Usage:
    $ docker run -it --rm \
        -e "URL=http://cyclone-server:7099/apis/v1alpha1/workflowruns/w1/artifacts?namespace=ns&stage=s1" \
        -e "METHOD=POST" \
        -e "HEADERS=Content-Type:multipart/form-data X-Tenant:cyclone" \
        -e "COMPRESS_FILE_NAME=artifacts.tar" \
        -e "FORM_FILE_KEY=file" \
        -e "CURL_EXTENSION=-i -v" \
        git-resource-http:latest <COMMAND>

      Supported commands are:
      - help Print out this help messages.
      - pull Send a http request to get sources to ${WORKDIR}/data, (Not implemented yet).
      - push Compress files in ${WORKDIR}/data to a file named $COMPRESS_FILE_NAME and 
        send a http request to upload the compressed file with form key $FORM_FILE_KEY 
        to a remote server.
        deep dive in implementation with pseudocode:
          $ tar -cvf $COMPRESS_FILE_NAME ${WORKDIR}/data
          $ curl $URL -H $HEADERS -X $METHOD -F "$FORM_FILE_KEY=@$COMPRESS_FILE_NAME" $CURL_EXTENSION

      Arguments:
      - URL [Required] URL of the remote server, for the moment, only HTTP is supported.
      - METHOD [Required] Http request command, like 'GET', 'POST', 'PUT', etc.
      - HEADERS [Optional] Http request headers, we can use this to pass some auth infomation. Multiple
        headers please separate with blank, for example: HEADERS="X-Tenant:devops X-User:cyclone".
      - COMPRESS_FILE_NAME [Optional] This tool will compress the output files to one file 
        named with this argument, by default is 'artifact.tar'.
      - FORM_FILE_KEY [Optional] The key of curl '-F' option (Specify HTTP multipart POST data) used to 
        upload files, by default is 'file'.
      - CURL_EXTENTION [Optional] Some other extent options of cURL
      - FIND_OPTIONS [Optional] is only used in output http resources. We will pass the FIND_OPTIONS to Linux
        command "find" to find files in the ${WORKDIR}/data/${DATA_SUBDIR} folder, and then we will tar and push them.
        E.g. ". -path './output' -name '*.jar'" will populate the command "find . -path './output' -name *.jar".
        Default value is ". -name '*'".
      - DATA_SUBDIR [Optional] is only used in output http resources. If DATA_SUBDIR is not empty, will find
        output resources in the ${WORKDIR}/data/${DATA_SUBDIR} folder.

      Notes:
      - This tool will replease the following strings, built-in environments set by workflow
       controller, in URL with corresponding real value.
        - ${METADATA_NAMESPACE}
        - ${WORKFLOW_NAME}
        - ${STAGE_NAME}
        - ${WORKFLOWRUN_NAME}
END
)

if [[ $# != 1 ]]; then
    echo "$USAGE"
    exit 1
fi
COMMAND=$1

# Check whether environment variables are set.
if [ -z ${WORKDIR} ]; then echo "WORKDIR is unset"; exit 1; fi
if [ -z ${URL} ]; then echo "URL is unset"; exit 1; fi
if [ -z ${METHOD} ]; then echo "METHOD is unset"; exit 1; fi

# replease string '${METADATA_NAMESPACE}' '${WORKFLOW_NAME}' '${STAGE_NAME}' '${WORKFLOWRUN_NAME}' in URL with corresponding real value
URL=${URL//'${METADATA_NAMESPACE}'/${METADATA_NAMESPACE}}
URL=${URL//'${WORKFLOW_NAME}'/${WORKFLOW_NAME}}
URL=${URL//'${STAGE_NAME}'/${STAGE_NAME}}
URL=${URL//'${WORKFLOWRUN_NAME}'/${WORKFLOWRUN_NAME}}

wrapPush() {
    if [ -z ${COMPRESS_FILE_NAME} ]; then COMPRESS_FILE_NAME="artifact.tar"; fi
    if [ -z ${FORM_FILE_KEY} ]; then FORM_FILE_KEY="file"; fi
    if [ -z ${FIND_OPTIONS} ]; then FIND_OPTIONS=". -name '*'"; fi

    cd ${WORKDIR}/data/${DATA_SUBDIR}
    mkdir -p ${WORKDIR}/__output_resources;

    echo "Start to find and copy files: find ${FIND_OPTIONS} -exec cp --parents {} ${WORKDIR}/__output_resources \;"
    eval "find ${FIND_OPTIONS} -exec cp -v --parents {} ${WORKDIR}/__output_resources \;"

    if [ -z "$(ls -A "${WORKDIR}/__output_resources")" ]; then
       echo "No files should be sent, exit."
       exit 0
    fi

    echo "Start to compress files under ${WORKDIR}/__output_resources into file ${COMPRESS_FILE_NAME}"
    cd ${WORKDIR}/__output_resources
    tar -cvf ${WORKDIR}/${COMPRESS_FILE_NAME} ./*
    cd ${WORKDIR}

    for header in ${HEADERS}; do
        headerString="-H ${header} ${headerString}"
    done

    echo "Start to upload file"
    echo "curl -v ${URL} ${headerString} -X ${METHOD} -F \"${FORM_FILE_KEY}=@${COMPRESS_FILE_NAME}\" ${CURL_EXTENSION}"
    
    status_code=$(curl --write-out %{http_code} --silent --output /dev/null ${URL} ${headerString} -X ${METHOD} -F "${FORM_FILE_KEY}=@${COMPRESS_FILE_NAME}" ${CURL_EXTENSION})
    if [[ "$status_code" -ne 201 ]] ; then
        echo "Upload files error, status code: $status_code"
        exit 1
    else
        exit 0
    fi
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
        echo "Not implemented yet"
        ;;
    push )
        wait_ok
        ls -la ${WORKDIR}/data
        wrapPush
        ;;
    * )
        echo "$USAGE"
        ;;
esac