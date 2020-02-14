#!/bin/bash

template=$(cat <<-END
{
  "total": "%s",
  "used": "%s",
  "items": {%s}
}
END
)

function join_by {
  local d=$1;
  shift;
  echo -n "$1";
  shift;
  printf "%s" "${@/#/$d}";
}

function usage {
  cd $1
  total=$(df -h "$1" | grep "$1" | awk '{ print $(NF-4) }')
  used=$(df -h "$1" | grep "$1" | awk '{ print $(NF-3) }')

  if [ -z "$(ls -A .)" ]; then
    printf "$template" "$total" "$used" ""
  else
    items=$(du -sh -- * | awk '{ printf("\"%s\":\"%s\"\n", $2, $1) }')
    jsonItems=$(join_by , "${items[@]}")
    printf "$template" "$total" "$used"  "$jsonItems"
  fi
}

echo "Report to $REPORT_URL every $HEARTBEAT_INTERVAL seconds."
while [ true ];
do
  echo -e "[`date '+%Y-%m-%d %H:%M:%S'`]: Start to get usage ..."
  json=$(usage "/pvc-data")
  echo -e "[`date '+%Y-%m-%d %H:%M:%S'`]: Finished get usage ..."
  wget -q -O /dev/null --header="Content-Type:application/json" --header="X-Namespace:$NAMESPACE" --post-data="$json" $REPORT_URL;
  sleep $HEARTBEAT_INTERVAL;
done
