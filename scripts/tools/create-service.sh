#!/bin/bash
#
# Usage: create-service-version.sh <username>
#
# Other alternative github repositories for test:
#   https://github.com/dweinstein/dockerfile-androguard

if [[ $# -ne 1 ]]; then
	echo "Arguments error."
	echo "Usage: create-service.sh <username>"
	exit 1
fi

echo "Using ${1} as username."

curl -sS -X POST -H "Content-Type:application/json" -d "{
   \"name\": \"test-service\",
   \"username\": \"${1}\",
   \"description\": \"This is a test-service\",
   \"yaml_config_name\": \"caicloud.yml\",
   \"repository\": {
     \"url\": \"https://github.com/tutumcloud/hello-world\",
     \"vcs\": \"git\"
   }
}" "http://localhost:7099/api/v0.1/fake-user-id/services"
