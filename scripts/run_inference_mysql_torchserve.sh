#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_run_inference_torchserve)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_rest_tensorflow not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

TORCHSERVE_PORT=${TORCHSERVE_PORT:-8080}

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
  for NUM_WORKERS in 1 16 32 48 64 80 96 112 128 144 160; do
    for RUN in 1 2 3; do
      NUM_INFERENCES_IN=$NUM_INFERENCES
      if [[ "${NUM_WORKERS}" == "1" ]]; then
        NUM_INFERENCES_IN=$((${NUM_INFERENCES} / 10))
      fi
      FILENAME_SUFFIX=torchserve_ref_mysql_${REFERENCE_DATA}_${OUTPUT_NAME_SUFIX}_run_${RUN}_workers_${NUM_WORKERS}_rate_${RATE_LIMIT}.txt
      echo "Benchmarking inference performance with reference data set to: ${REFERENCE_DATA} and model name ${MODEL_NAME}"
      echo "\t\tSaving files with file suffix: ${FILENAME_SUFFIX}"

      # benchmark inference performance
      # make sure you're on the root project folder
      cd $GOPATH/src/github.com/RedisAI/aibench
      cat ${DATA_FILE} |
        ${EXE_FILE_NAME} \
          -workers=${NUM_WORKERS} \
          -print-responses=${PRINT_RESPONSES} \
          -burn-in=${QUERIES_BURN_IN} -max-queries=${NUM_INFERENCES_IN} \
          -print-interval=0 -reporting-period=1000ms \
          -limit-rps=${RATE_LIMIT} \
          -debug=${DEBUG} \
          -enable-reference-data-mysql=${REFERENCE_DATA} \
          -mysql-host="${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/" \
          -ignore-errors=true \
          -torchserve-host=${MODELSERVER_HOST}:${TORCHSERVE_PORT} \
          -redis-host=${DATABASE_HOST}:${DATABASE_PORT} \
          -output-file-stats-hdr-response-latency-hist=~/HIST_${FILENAME_SUFFIX} \
          2>&1 | tee ~/RAW_${FILENAME_SUFFIX}

      echo "Sleeping: $SLEEP_BETWEEN_RUNS"
      sleep ${SLEEP_BETWEEN_RUNS}
    done
  done
done
