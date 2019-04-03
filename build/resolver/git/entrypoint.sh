#!/bin/bash
set -e

USAGE=$(cat <<-END
    This tool is used to resolve git resources, here git resource
    stands for a revision in a git repository.

    Usage:
        $ docker run -it --rm \\
            -e GIT_URL=https://github.com/caicloud/cyclone.git \\
            -e GIT_REVISION=master \\
            -e GIT_TOKEN=xxxx \\
            -e PULL_POLICY=IfNotPresent \\
            git-resource-resolver:latest <COMMAND>

     Supported commands are:
     - help Print out this help messages.
     - pull Pull git source to $WORKDIR/data, "/workspace/data" by default.
     - push Push git source to remote git server. (Not implemented yet)

     Arguments:
     - GIT_URL [Required] URL of the git repository, for the moment, only HTTP/
       HTTPS are supported.
     - GIT_REVISION [Required] Revision of the source code. It has two different
       format. a) Single revision, such as branch 'master', tag 'v1.0'; b). Composite
       such as pull requests, 'develop:master' indicates merge 'develop' branch to
       'master'. For GitHub, pull requests can use the single revision form, such as
       'refs/pull/1/merge', but for Gitlab, composite revision is necessary, such as
       'refs/merge-requests/1/head:master'.
     - GIT_TOKEN [Optional] For public repository, no need provide this token, but for
       private repository, this should be provided.
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
if [ -z ${WORKDIR+x} ]; then echo "WORKDIR is unset"; exit 1; fi
if [ -z ${GIT_URL+x} ]; then echo "GIT_URL is unset"; exit 1; fi
if [ -z ${GIT_REVISION+x} ]; then echo "GIT_REVISION is unset"; exit 1; fi
if [ -z ${GIT_TOKEN+x} ]; then echo "WARN: GIT_TOKEN is unset"; fi

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

# Revision can be in two different format:
# - Single revision. For example, 'master', 'develop' as branch, 'v1.0' as tag, etc.
# - Composite revision. For example, 'develop->master', it means merge branch 'develop' to 'master'. It's used in merge
#   request in GitLab.
# This function parses the composite revision to get the source and target branches. For example,
#   'develop->master' --> ['develop', 'master']
parseRevision() {
    SOURCE_BRANCH=${GIT_REVISION%%:*}
    TARGET_BRANCH=${GIT_REVISION##*:}
}
parseRevision

pull() {
    # If data existed and pull policy is IfNotPresent, perform incremental pull.
    if [ -e $WORKDIR/data ] && [ ${PULL_POLICY:=Always} == "IfNotPresent" ]; then
        cd $WORKDIR/data
        # Ensure existed data come from the git repo
        git remote -v | grep ${GIT_URL##*//} || {
            echo "Existed data not a valid git repo for ${GIT_URL##*//}"
            exit 1
        }

        echo "Fetch $GIT_REVISION from origin"
        git fetch -v origin $GIT_REVISION
        git checkout FETCH_HEAD
    else
        if [ -e $WORKDIR/data ]; then
            echo "Clean old data ($WORKDIR/data) when pull policy is Always"
            rm -rf $WORKDIR/data
        fi
        cd $WORKDIR

        # Add token to url if provided and clone git repo
        if [ ! -z ${GIT_TOKEN+x} ]; then
            GIT_URL=${GIT_URL/\/\//\/\/oauth2:${GIT_TOKEN}@}
        fi

        if [[ "${SOURCE_BRANCH}" == "${TARGET_BRANCH}" ]]; then
            echo "Clone $SOURCE_BRANCH..."
            git clone -v -b master --single-branch --recursive ${GIT_URL} data
            cd data
            git fetch origin $SOURCE_BRANCH
            git checkout -qf FETCH_HEAD
        else
            echo "Merge $SOURCE_BRANCH to $TARGET_BRANCH..."
            git clone -v -b $TARGET_BRANCH --single-branch --recursive ${GIT_URL} data
            cd data
            git config user.email "cicd@cyclone.dev"
            git config user.name "cicd"
            git fetch origin $SOURCE_BRANCH
            git merge FETCH_HEAD --no-ff --no-commit
        fi

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