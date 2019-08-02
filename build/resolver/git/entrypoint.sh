#!/bin/bash
set -e

USAGE=$(cat <<-END
    This tool is used to resolve git resources, here git resource
    stands for a revision in a git repository.

    Usage:
        $ docker run -it --rm \\
            -e SCM_URL=https://github.com/caicloud/cyclone.git \\
            -e SCM_REPO=caicloud/cyclone \\
            -e SCM_REVISION=master \\
            -e SCM_AUTH=xxxx \\
            -e PULL_POLICY=IfNotPresent \\
            git-resource-resolver:latest <COMMAND>

     Supported commands are:
     - help Print out this help messages.
     - pull Pull git source to $WORKDIR/data, "/workspace/data" by default.
     - push Push git source to remote git server. (Not implemented yet)

     Arguments:
     - SCM_URL [Required] URL of the git repository, for the moment, only HTTP/
       HTTPS are supported.
     - SCM_REPO [Optional] Repo of the source code, if URL doesn't include repo,
       repo can be given here.
     - SCM_REVISION [Required] Revision of the source code. It has two different
       format. a) Single revision, such as branch 'master', tag 'v1.0'; b). Composite
       such as pull requests, 'develop:master' indicates merge 'develop' branch to
       'master'. For GitHub, pull requests can use the single revision form, such as
       'refs/pull/1/merge', but for Gitlab, composite revision is necessary, such as
       'refs/merge-requests/1/head:master'.
     - SCM_AUTH [Optional] For public repository, no need provide auth, but for
       private repository, this should be provided. Auth here supports 2 different formats:
       a. <user>:<password>
       b. <token>
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
if [ -z ${SCM_REVISION} ]; then echo "SCM_REVISION is unset"; exit 1; fi
if [ -z ${SCM_AUTH} ]; then echo "WARN: SCM_AUTH is unset"; fi

# If SCM_REPO is provided, embed it to SCM_URL
if [ ! -z ${SCM_REPO} ]; then
    SCM_URL=${SCM_URL%/}/${SCM_REPO}.git
fi

# Lock folder for the WorkflowRun.
PULLING_LOCK=$WORKDIR/${WORKFLOWRUN_NAME}-pulling.lock

releaseLock() {
    if [ -d "$PULLING_LOCK" ]; then
        echo "Remove the created lock folder"
        rm -rf $PULLING_LOCK
    fi
}

# Make sure the lock file is deleted if created by this resolver.
trap releaseLock EXIT

urlencode() {
    local string="${1}"
    local strlen=${#string}
    local encoded=""
    local pos c o

    for (( pos=0 ; pos<strlen ; pos++ )); do
        c=${string:$pos:1}
        case "$c" in
            [-_.~a-zA-Z0-9] ) o="${c}" ;;
            * )               printf -v o '%%%02x' "'$c"
         esac
        encoded+="${o}"
    done
    echo "${encoded}"
}

# Add credential to URL
modifyURL() {
    local URL=$1
    local AUTH=$2
    LEFT=${AUTH%%:*}
    RIGHT=${AUTH##*:}
    if [[ "$LEFT" == "$AUTH" ]]; then
        AUTH="oauth2:$(urlencode "$AUTH")"
    else
        AUTH="$(urlencode "$LEFT"):$(urlencode "$RIGHT")"
    fi

    # If URL contains '@', for example: http://root@192.168.21.97/scm/foobar.git, change
    # url to http://root:<auth>@192.168.21.97/scm/foobar.git
    if [[ "${URL%%@*}" != "${URL}" ]]; then
        local encoded=$(urlencode "${2}")
        echo ${URL/@/:${encoded}@}
    else
        echo ${URL/\/\//\/\/${AUTH}@}
    fi
}

wrapPull() {
    # If there is already data, we should just wait for the data to be ready.
    if [ -e $WORKDIR/data ]; then
        echo "Found data, wait it to be ready..."

        # If there already be data, then we just wait the data to be ready
        # by checking the lock folder. If the lock folder exists, it means other
        # stage is pulling the data.
        while [ -d $PULLING_LOCK ]
        do
            sleep 3
        done
    else
        echo "Trying to acquire lock and pull resource"
        failed=$(mkdir $PULLING_LOCK > /dev/null 2>&1 || echo fail)
        if [[ $failed != "fail" ]]; then
            echo "Got the lock, start to pulling..."
            pull
        else
            echo "Failed to get the lock, wait others to finish pulling..."
            while [ -d $PULLING_LOCK ]
            do
                sleep 3
            done
        fi
    fi

    # Write commit id to output file, which will be collected by Cyclone
    cd $WORKDIR/data
    echo "Collect commit id to result file /__result__ ..."
    echo "LastCommitID:`git log -n 1 --pretty=format:"%H"`" >> /__result__;
    cat /__result__;
}

# Revision can be in two different format:
# - Single revision. For example, 'master', 'develop' as branch, 'v1.0' as tag, etc.
# - Composite revision. For example, 'develop:master', it means merge branch 'develop' to 'master'. It's used in merge
#   request in GitLab.
# This function parses the composite revision to get the source and target branches. For example,
#   'develop:master' --> ['develop', 'master']
parseRevision() {
    SOURCE_BRANCH=${SCM_REVISION%%:*}
    TARGET_BRANCH=${SCM_REVISION##*:}
}
parseRevision

pull() {
    git config --global http.sslVerify false
    git config --global http.postBuffer 500M

    # If data existed and pull policy is IfNotPresent, perform incremental pull.
    if [ -e $WORKDIR/data ] && [ ${PULL_POLICY:=Always} == "IfNotPresent" ]; then
        cd $WORKDIR/data
        # Ensure existed data come from the git repo
        git remote -v | grep ${SCM_URL##*//} || {
            echo "Existed data not a valid git repo for ${SCM_URL##*//}"
            exit 1
        }

        echo "Fetch $SCM_REVISION from origin"
        git fetch -v --depth=1 origin $SCM_REVISION
        git checkout FETCH_HEAD
    else
        if [ -e $WORKDIR/data ]; then
            echo "Clean old data ($WORKDIR/data) when pull policy is Always"
            rm -rf $WORKDIR/data
        fi
        cd $WORKDIR

        # Add auth to url if provided and clone git repo. If SCM_AUTH is in format '<user>:<password>', then url
        # encode each part of it and get '<encoded_user>:<encoded_password>'. If SCM_AUTH is in format '<token>',
        # give it a 'oauth2:' prefix to get 'oauth2:<encoded_token>'.
        if [ ! -z ${SCM_AUTH+x} ]; then
            SCM_URL=$( modifyURL $SCM_URL $SCM_AUTH )
        fi

        if [[ "${SOURCE_BRANCH}" == "${TARGET_BRANCH}" ]]; then
            echo "Clone $SOURCE_BRANCH..."
            git clone -v -b master --depth=1 --single-branch --recursive ${SCM_URL} data
            cd data
            git fetch --depth=1 origin $SOURCE_BRANCH
            git checkout -qf FETCH_HEAD
        else
            echo "Merge $SOURCE_BRANCH to $TARGET_BRANCH..."
            git clone -v -b $TARGET_BRANCH --depth=1 --single-branch --recursive ${SCM_URL} data
            cd data
            git config user.email "cicd@cyclone.dev"
            git config user.name "cicd"
            # If the fetch depth is too small, the merge command will fail with:
            #    'fatal: refusing to merge unrelated histories'
            # And we assume 30 is enough.
            git fetch --depth=30 origin $SOURCE_BRANCH
            git merge FETCH_HEAD --no-ff --no-commit
        fi
    fi

    cd $WORKDIR/data
    ls -al
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