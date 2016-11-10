#!/bin/bash
#
# Usage: create-service-version.sh <username>

if [[ $# -ne 2 ]]; then
	echo "Arguments error."
	echo "Usage: create-service-version.sh <username> <operation>"
	exit 1
fi

echo "Using ${1} as username."

# Create a new service.
OUTPUT=\
"$(curl -sS -X POST -H "Content-Type:application/json" -d "{
   \"name\": \"test-service\",
   \"username\": \"${1}\",
   \"description\": \"This is a test-service\",
   \"yaml_config_name\": \"caicloud.yml\",
   \"repository\": {
     \"url\": \"https://github.com/caicloud/toy-dockerfile\",
     \"vcs\": \"git\"
   },
   \"profile\": {
     \"setting\": \"sendwhenfinished\",
   }
}" "http://localhost:7099/api/v0.1/fake-user-id/services")"

# Get the service id.
SERVICEID="$(echo ${OUTPUT} | python -c "import json,sys;obj=json.load(sys.stdin);print obj[\"service_id\"]")"

# Async, so need to wait clone operation to finish.
sleep 15s

# Create a new version.
curl -sS -X POST -H "Content-Type:application/json" -d "{
   \"name\": \"v0.1.1\",
   \"description\": \"just for test\",
   \"service_id\": \"${SERVICEID}\",
   \"operation\": \"${2}\"
}" "http://localhost:7099/api/v0.1/fake-user-id/versions"
