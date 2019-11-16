#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_run_inference_flask_tensorflow)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_rest_tensorflow not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

DATA_FILE_NAME=${DATA_FILE_NAME:-aibench_generate_data-creditcard-fraud.dat.gz}
MAX_QUERIES=${MAX_QUERIES:-0}
RESTAPI_PORT=${RESTAPI_PORT:-8000}
QUERIES_BURN_IN=${QUERIES_BURN_IN:-10}

# Rate limit? if greater than 0 rate is limited.
RATE_LIMIT=${RATE_LIMIT:-0}

# output name
OUTPUT_NAME_SUFIX=${OUTPUT_NAME_SUFIX:-""}

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Cannot find data file ${DATA_FILE}"
  exit 1
fi

for REFERENCE_DATA in "false"; do
  if [[ "${REFERENCE_DATA}" == "false" ]]; then
    MODEL_NAME=$MODEL_NAME_NOREFERENCE
  fi
  echo "Benchmarking inference performance with reference data set to: ${REFERENCE_DATA} and model name ${MODEL_NAME}"

  # benchmark inference performance
  # make sure you're on the root project folder
  redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} config resetstat
  cd $GOPATH/src/github.com/RedisAI/aibench
  cat ${BULK_DATA_DIR}/aibench_generate_data-creditcard-fraud.dat.gz |
    gunzip |
    ${EXE_FILE_NAME} \
      -workers=${NUM_WORKERS} \
      -burn-in=${QUERIES_BURN_IN} -max-queries=${MAX_QUERIES} \
      -print-interval=0 -reporting-period=1000ms \
      -limit-rps=${RATE_LIMIT} \
      -debug=${DEBUG} \
      -enable-reference-data=${REFERENCE_DATA} \
      -output-file-stats-hdr-response-latency-hist=rest_tensorflow_referencedata_${REFERENCE_DATA}_hdr_${OUTPUT_NAME_SUFIX}_${NUM_WORKERS}_workers_${RATE_LIMIT}.txt \
      -restapi-host=${DATABASE_HOST}:${RESTAPI_PORT} \
      -redis-host=${DATABASE_HOST}:${DATABASE_PORT} 2>&1 | tee ~/rest_tensorflow_referencedata_${REFERENCE_DATA}_results_${OUTPUT_NAME_SUFIX}_${NUM_WORKERS}_workers_${RATE_LIMIT}.txt

  redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats

done
