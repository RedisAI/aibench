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

for REFERENCE_DATA in "true"; do
  if [[ "${REFERENCE_DATA}" == "false" ]]; then
    MODEL_NAME=$MODEL_NAME_NOREFERENCE
  fi
  # we overload the NUM_WORKERS here for the official benchmark
  for NUM_WORKERS in 16 32 48 64 80 96 112 128 144 160; do
    for RUN in 1 2 3; do
      FILENAME_SUFFIX=redisai_ref_${REFERENCE_DATA}_${OUTPUT_NAME_SUFIX}_run_${RUN}_workers_${NUM_WORKERS}_rate_${RATE_LIMIT}.txt
      echo "Benchmarking inference performance with reference data set to: ${REFERENCE_DATA} and model name ${MODEL_NAME}"
      echo "\t\tSaving files with file suffix: ${FILENAME_SUFFIX}"

      # benchmark inference performance
      # make sure you're on the root project folder
      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} config resetstat
      cd $GOPATH/src/github.com/RedisAI/aibench
      cat ${DATA_FILE} |
        ${EXE_FILE_NAME} \
          -workers=${NUM_WORKERS} \
          -burn-in=${QUERIES_BURN_IN} -max-queries=${MAX_QUERIES} \
          -print-interval=0 -reporting-period=1000ms \
          -limit-rps=${RATE_LIMIT} \
          -debug=${DEBUG} \
          -enable-reference-data=${REFERENCE_DATA} \
          -restapi-host=${DATABASE_HOST}:${RESTAPI_PORT} \
          -redis-host=${DATABASE_HOST}:${DATABASE_PORT} \
          -output-file-stats-hdr-response-latency-hist=~/HIST_${FILENAME_SUFFIX} \
          2>&1 | tee ~/RAW_${FILENAME_SUFFIX}

      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats
      echo "Sleeping: $SLEEP_BETWEEN_RUNS"
      sleep ${SLEEP_BETWEEN_RUNS}
    done
  done
done
