#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_run_inference_redisai)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_redisai not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

DATA_FILE_NAME=${DATA_FILE_NAME:-aibench_generate_data-creditcard-fraud.dat.gz}
QUERIES_BURN_IN=${QUERIES_BURN_IN:-10}

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Cannot find data file ${DATA_FILE}"
  exit 1
fi

# benchmark inference performance
# make sure you're on the root project folder
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} config resetstat
cd $GOPATH/src/github.com/RedisAI/aibench
cat ${BULK_DATA_DIR}/aibench_generate_data-creditcard-fraud.dat.gz |
  gunzip |
  ${EXE_FILE_NAME} \
    -workers ${NUM_WORKERS} \
    -burn-in ${QUERIES_BURN_IN} -max-queries ${NUM_INFERENCES} \
    -print-interval 0 -reporting-period 1000ms \
    -model ${MODEL_NAME} \
    -host redis://${DATABASE_HOST}:${DATABASE_PORT}
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats

