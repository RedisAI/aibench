#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_load_data)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_load_data not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Cannot find data file ${DATA_FILE}"
  exit 1
fi

# load the reference data in BLOB format
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
cat ${DATA_FILE} |
  ${EXE_FILE_NAME} \
    -reporting-period 1000ms \
    -set-blob=true \
    -use-redis=false \
    -use-mysql=true \
    -max-inserts=${NUM_INFERENCES} \
    -mysql-host="perf:perf@tcp(${MYSQL_HOST}:${MYSQL_PORT})/test" \
    -redis-host="redis://${DATABASE_HOST}:${DATABASE_PORT}" \
    -workers=${NUM_WORKERS} -pipeline=1000

redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats
