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
          $ curl $URL -H $HEADERS -X $METHOD -F "$FORM_FILE_KEY@$COMPRESS_FILE_NAME" $CURL_EXTENSION

      Arguments:
      - URL [Required] URL of the remote server, for the moment, only HTTP is supported.
      - METHOD [Required] Http request command, like 'GET', 'POST', 'PUT', etc.
      - HEADERS [Optional] Http request headers, we can use this to pass some auth infomation. Multiple
        headers please separate with blank, for example: HEADERS="X-Tenant:devops X-User:cyclone".
      - COMPRESS_FILE_NAME [Optional] This tool will compress the output files to one file 
        named with this argument, by default is 'artifact.tar'.
      - FORM_FILE_KEY [Optional] The key of curl '-F' option (Specify HTTP multipart POST data) used to 
        upload files, by default is 'file'.
      - CURL_EXTENSION [Optional] Some other extent options of cURL

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

    echo "Start to compress files under ${WORKDIR}/data into file ${COMPRESS_FILE_NAME}"
    cd ${WORKDIR}/data
    tar -cvf ${WORKDIR}/${COMPRESS_FILE_NAME} *
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
        ls -la /${WORKDIR}/data
        wrapPush
        ;;
    * )
        echo "$USAGE"
        ;;
esac