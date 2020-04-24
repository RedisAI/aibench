#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_run_inference_tensorflow_serving)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_tensorflow_serving not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

TFX_MODEL_VERSION=${TFX_MODEL_VERSION:-2}
TFX_PORT=${TFX_PORT:-8500}

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
        -workers=${NUM_WORKERS} \
          -burn-in=${QUERIES_BURN_IN} -max-queries=${MAX_QUERIES} \
          -print-interval=0 -reporting-period=1000ms \
          -limit-rps=${RATE_LIMIT} \
          -enable-reference-data=${REFERENCE_DATA} \
          -model=${MODEL_NAME} -model-version=${TFX_MODEL_VERSION} \
          -tensorflow-serving-host=${DATABASE_HOST}:${TFX_PORT} \
          -redis-host=${DATABASE_HOST}:${DATABASE_PORT} \
          -output-file-stats-hdr-response-latency-hist=~/HIST_${FILENAME_SUFFIX} \
          2>&1 | tee ~/RAW_${FILENAME_SUFFIX}

      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats
      echo "Sleeping: $SLEEP_BETWEEN_RUNS"
      sleep ${SLEEP_BETWEEN_RUNS}
    done
  done
done
