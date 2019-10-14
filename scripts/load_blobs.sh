#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_load_data)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_load_data not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

DATA_FILE_NAME=${DATA_FILE_NAME:-aibench_generate_data-creditcard-fraud.dat.gz}

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
  gunzip |
  ${EXE_FILE_NAME} \
    -reporting-period 1000ms \
    -set-blob=true \
    -host redis://${DATABASE_HOST}:${DATABASE_PORT} \
    -workers ${NUM_WORKERS} -pipeline 1000 #-max-inserts 10000

redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats
