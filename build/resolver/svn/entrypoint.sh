#!/bin/bash
set -e
shopt -s extglob

USAGE=$(cat <<-END
    This tool is used to resolve svn resources, here git resource
    stands for a revision in a svn repository.

    Usage:
        $ docker run -it --rm \\
            -e SCM_URL=http://192.168.21.97/svn \\
            -e SCM_REVISION=996 \\
            -e SCM_USER=root \\
            -e SCM_PWD=pwd \\
            svn-resource-resolver:latest <COMMAND>

     Supported commands are:
     - help Print out this help messages.
     - pull Pull svn source to $WORKDIR/data, "/workspace/data" by default.
     - push Push svn source to remote svn server. (Not implemented yet)

     Arguments:
     - SCM_URL [Required] URL of the svn repository.
     - SCM_REVISION [Required] Revision of the source code. For example, "996".
     - SCM_USER [Required] User name of the svn server.
     - SCM_PWD [Required] Password for the user.
     - PULL_POLICY [Optional] Indicate whether pull resources when there already
       are old data, if set to IfNotPresent, will make use of the old data and
       perform incremental pull, otherwise old data would be removed.
END
)

if [[ $# != 1 ]]; then
    echo "$USAGE"
    exit 1
fi
COMMAND=$1

# Check whether environment variables are set.
if [ -z ${WORKDIR} ]; then echo "WORKDIR is unset"; exit 1; fi
if [ -z ${SCM_URL} ]; then echo "SCM_URL is unset"; exit 1; fi
if [ -z ${SCM_REVISION} ]; then echo "SCM_REVISION is unset"; fi
if [ -z ${SCM_USER} ]; then echo "SCM_REVISION is unset"; exit 1; fi
if [ -z ${SCM_PWD} ]; then echo "SCM_REVISION is unset"; exit 1; fi

# Lock file for the WorkflowRun.
PULLING_LOCK=$WORKDIR/${WORKFLOWRUN_NAME}-pulling.lock

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
    if [ -e $WORKDIR/data ]; then
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
        if [ $result == "ok" ]; then
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
    # If data existed and pull policy is IfNotPresent, perform incremental pull.
    if [ -e $WORKDIR/data ] && [ ${PULL_POLICY:=Always} == "IfNotPresent" ]; then
        cd $WORKDIR/data
        svn update -r ${SCM_REVISION}
    else
        if [ -e $WORKDIR/data ]; then
            echo "Clean old data ($WORKDIR/data) when pull policy is Always"
            rm -rf $WORKDIR/data
        fi
        cd $WORKDIR

        if [[ "${SCM_REVISION}" == "" ]]; then
            svn checkout --username ${SCM_USER} --password ${SCM_PWD} --non-interactive --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other --no-auth-cache ${SCM_URL} data
        else
            case ${SCM_REVISION} in
                +([0-9]))
                    svn checkout --username ${SCM_USER} --password ${SCM_PWD} --revision ${SCM_REVISION} --non-interactive --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other --no-auth-cache ${SCM_URL} data
                    ;;
                *)
                    svn checkout --username ${SCM_USER} --password ${SCM_PWD} --non-interactive --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other --no-auth-cache ${SCM_URL}/${SCM_REVISION} data
                    ;;
            esac
        fi

        cd data
        ls -al
    fi
}

case $COMMAND in
    pull )
        wrapPull
        ;;
    push )
        echo "Not implemented yet"
        ;;
    * )
        echo "$USAGE"
        ;;
esac