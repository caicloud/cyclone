#!/bin/sh
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

     Environment variables GIT_URL, GIT_REVISION must be set. Only HTTPS is
     supported now for GIT_URL. And revision supports branch and tag, but not
     commit id. PULL_POLICY indicates whether pull resources when there already
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

cd $WORKDIR/data

pull() {
    # Get repo name from the url. For example,
    # https://xxx@github.com/caicloud/cyclone.git --> cyclone
    TMP=${GIT_URL##*/}
    REPO=${TMP%.git}

    # If data existed and pull policy is IfNotPresent, perform incremental pull.
    if [ -e $WORKDIR/data/$REPO ] && [ ${PULL_POLICY:=Always} == "IfNotPresent" ]; then
        # Ensure existed data come from the git repo
        cd ./$REPO
        git remote -v | grep ${GIT_URL##*//} || {
            echo "Existed data not a valid git repo for ${GIT_URL##*//}"
            exit 1
        }

        echo "Fetch $GIT_REVISION from origin"
        git fetch -v origin $GIT_REVISION
        git checkout FETCH_HEAD
    else
        if [ -e $WORKDIR/data/$REPO ]; then
            echo "Clean old data ($WORKDIR/data/$REPO) when pull policy is Always"
            rm -rf $WORKDIR/data/$REPO
        fi

        # Add token to url if provided and clone git repo
        if [ -z ${GIT_TOKEN+x} ]; then
            git clone -v -b $GIT_REVISION --single-branch ${GIT_URL}
        else
            git clone -v -b $GIT_REVISION --single-branch ${GIT_URL/\/\//\/\/${GIT_TOKEN}@}
        fi
    fi
}

case $COMMAND in
    pull )
        pull
        ;;
    push )
        echo "Not implemented yet"
        ;;
    * )
        echo "$USAGE"
        ;;
esac