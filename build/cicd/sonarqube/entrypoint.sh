#!/usr/bin/env bash
set -e

USAGE=$(cat <<-END
    This image is used to perform source code scanning with SonarQube.

    Usage:
        $ docker run -it --rm \\
            -e SERVER=http://192.168.21.96:9000 \\
            -e TOKEN=58a66fa93ee4a5efe5c8cea9351526efdb82b792 \\
            -e ENCODING=UTF-8 \\
            -e LANGUAGE=Go \\
            -e PROJECT_NAME=codescan \\
            -e PROJECT_KEY=codescan \\
            -e QUALITY_GATE=1 \\
            -e SOURCE_PATH=./ \\
            -e EXTENSION_PARAMETERS="-Dsonar.exclusions=vendor/***,node_modules/***" \\
            cyclone-cicd-sonarqube:latest

     Arguments:
     - SERVER [Required] URL of the SonarQube server
     - TOKEN [Required] Token of SonarQube for authentication
     - ENCODING [Optional] Encoding of the source files, e.g. UTF-8
     - LANGUAGE [Optional] Language of the source files
     - PROJECT_NAME [Required] Name of the project that will be displayed on the SonarQube web interface
     - PROJECT_KEY [Required] The project's unique key. Allowed characters are: letters, numbers, - , _ , . and : , with at least one non-digit
     - QUALITY_GATE [Optional] Quality gate enforces a quality policy by defining a set of Boolean conditions based on measure thresholds against which projects are measured.
     - SOURCE_PATH [Optional] Comma-separated paths to directories containing main source files.
     - EXTENSION_PARAMETERS [Optional] Bland-separated extension parameters to sonar scanner.
END
)
echo $1
if [[ $# == 1 && $1 == "help" ]]; then
    echo "$USAGE"
    exit 0
fi

# Check whether environment variables are set.
if [ -z ${SERVER} ]; then echo "SERVER is unset"; exit 1; fi
if [ -z ${TOKEN} ]; then echo "TOKEN is unset"; exit 1; fi
if [ -z ${PROJECT_NAME} ]; then echo "PROJECT_NAME is unset"; exit 1; fi
if [ -z ${PROJECT_KEY} ]; then echo "PROJECT_KEY is unset"; exit 1; fi

if [ -z ${ENCODING} ]; then echo "ENCODING is unset, set it to UTF-8"; ENCODING=UTF-8; fi
if [ -z ${SOURCE_PATH} ]; then echo "SOURCE_PATH is unset, set it to ./"; SOURCE_PATH=./ ; fi

# Trim suffix "/" for the server
SERVER=$(echo ${SERVER} | sed -e 's/\/$//')

# Create project if not exist
status=$(curl -I -u ${TOKEN}: ${SERVER}/api/components/show?component=${PROJECT_KEY} 2>/dev/null | head -n 1 | cut -d$' ' -f2)
if [[ $status == "404" ]]; then
    echo "Project with key ${PROJECT_KEY} not exist, create it..."
    curl -XPOST -u ${TOKEN}: "${SERVER}/api/projects/create?name=${PROJECT_NAME}&project=${PROJECT_KEY}"
elif [[ "$status" == "200" ]]; then
    echo "Project with key ${PROJECT_KEY} already exist, no need to create..."
else
    echo "Failed to check and create SonarQube project"
    exit 1;
fi

# Set quality gate, if not provided, use default quality gate: Sonar way, with id '1'
curl -XPOST -u ${TOKEN}: "${SERVER}/api/qualitygates/select?gateId=${QUALITY_GATE:-1}&projectKey=${PROJECT_KEY}" 2>/dev/null | grep errors && {
    echo "Set quality gate error, exit"
    exit 1
}

declare -A paramsMap=(
  ['sonar.sourceEncoding']="$ENCODING"
  ['sonar.projectName']="$PROJECT_NAME"
  ['sonar.projectKey']="$PROJECT_KEY"
  ['sonar.sources']="$SOURCE_PATH"
  ['sonar.projectBaseDir']="$SOURCE_PATH"
  ['sonar.host.url']="$SERVER"
  ['sonar.login']="$TOKEN"
)

if [[ $LANGUAGE = Java ]]; then
    paramsMap['sonar.sources']='./src/main'
    paramsMap['sonar.java.binaries']='./target/classes'
    paramsMap['sonar.junit.reportPaths']='./target/surefire-reports'
fi

if [[ -f ./sonar-project.properties ]]; then
    while read -r line; do
        # ignore the line starts with '#'
        if [[ $line = \#* ]] || [[ -z $line ]]; then
            continue
        fi
        IFS='=' read -r arg _rest <<< "$line"
        unset paramsMap["$arg"]
    done < ./sonar-project.properties
fi

case ${LANGUAGE} in
    Go)
        if [[ -z ${paramsMap['sonar.go.coverage.reportPaths']} ]]; then
            echo "Start to find go test reports."
            testReportFile=$(find . -name "coverage.out" | tr '\n' ',')
            if [[ -n ${testReportFile} ]]; then
                paramsMap['sonar.go.coverage.reportPaths']="$testReportFile"
            fi
        fi
        ;;
esac

params=''
for key in "${!paramsMap[@]}"; do
    params+=" -D${key}=${paramsMap[$key]}"
done

# Set sonar-scanner extension parameters
if [[ -n ${EXTENSION_PARAMETERS} ]]; then echo "EXTENSION_PARAMETERS: ${EXTENSION_PARAMETERS}"; params="$params $EXTENSION_PARAMETERS"; fi

echo "[DEBUG] sonar-scanner ${params/$TOKEN/**********} -X"
# Scan the source files
sonar-scanner $params -X;

# Wait for the scan task completed
echo "Wait for scan task completed..."
while true; do
    ceTaskUrl=$(cat ./.scannerwork/report-task.txt | grep ceTaskUrl | cut -c11-)
    taskStatus=$(curl -u ${TOKEN}: $ceTaskUrl 2>/dev/null | tr ',' '\n' | grep "status" | awk -F: '{ print $2 }')
    echo $taskStatus | grep -v FAILED | grep -v SUCCESS || break;
    echo "Check in 3 seconds..."
    sleep 3;
done;
echo "Scan task completed~"

# Write result to output file, which will be collected by Cyclone
echo "Collect result to result file /__result__ ..."
echo "detailURL:${SERVER}/dashboard?id=${PROJECT_KEY}" >> /__result__;
# Can reference measures result in 'result.example.json' file in current directory
measures=$(curl -XPOST -u ${TOKEN}: "${SERVER}/api/measures/component?additionalFields=periods&component=${PROJECT_KEY}&metricKeys=reliability_rating,sqale_rating,security_rating,coverage,duplicated_lines_density,quality_gate_details" 2>/dev/null)
selected=$(echo $measures | jq -c .component.measures)
echo $selected | jq
echo "measures:${selected}" >> /__result__;

# Determine success or failure
qualityGateValue=$(echo $selected | jq '.[] | select(.metric=="quality_gate_details") | .value')
qualityGateValueWithoutQuote=$(echo $qualityGateValue | sed -e 's/\\//g' -e 's/^"//' -e 's/"$//')
qualityGateLevel=$(echo $qualityGateValueWithoutQuote | jq .level | sed -e 's/^"//' -e 's/"$//')
echo "quality gate level: ${qualityGateLevel}"

# qualityGateLevel: OK, ERROR, WARN, NONE
if [[ $qualityGateLevel = "ERROR" || $qualityGateLevel = "NONE" ]]; then
    exit 1;
else
    exit 0;
fi