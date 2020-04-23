#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_run_inference_redisai)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_redisai not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

QUERIES_BURN_IN=${QUERIES_BURN_IN:-10}

# Rate limit? if greater than 0 rate is limited.
RATE_LIMIT=${RATE_LIMIT:-0}

# output name
OUTPUT_NAME_SUFIX=${OUTPUT_NAME_SUFIX:-""}
PRINT_RESPONSES=${PRINT_RESPONSES:-false}

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Cannot find data file ${DATA_FILE}"
  exit 1
fi
#"false"
for REFERENCE_DATA in "true"; do
  if [[ "${REFERENCE_DATA}" == "false" ]]; then
    MODEL_NAME=$MODEL_NAME_NOREFERENCE
  fi
  echo "Benchmarking inference performance with reference data set to: ${REFERENCE_DATA} and model name ${MODEL_NAME}"
  # benchmark inference performance
  # make sure you're on the root project folder
  redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} config resetstat
  cd $GOPATH/src/github.com/RedisAI/aibench
  cat ${DATA_FILE} |
    ${EXE_FILE_NAME} \
      -model=${MODEL_NAME} \
      -model-filename=./tests/models/tensorflow/creditcardfraud.pb \
      -workers=${NUM_WORKERS} \
      -print-responses=${PRINT_RESPONSES} \
      -burn-in=${QUERIES_BURN_IN} -max-queries=${NUM_INFERENCES} \
      -print-interval=0 -reporting-period=1000ms \
      -limit-rps=${RATE_LIMIT} \
      -debug=${DEBUG} \
      -use-dag=true \
      -cluster-mode \
      -enable-reference-data=${REFERENCE_DATA} \
      -host=${DATABASE_HOST} \
      -port=${DATABASE_PORT} \
      -output-file-stats-hdr-response-latency-hist=redisai_referencedata_${REFERENCE_DATA}_hdr_${OUTPUT_NAME_SUFIX}_${NUM_WORKERS}_workers_${RATE_LIMIT}.txt \
      2>&1 | tee ~/redisai_referencedata_${REFERENCE_DATA}_results_${OUTPUT_NAME_SUFIX}_${NUM_WORKERS}_workers_${RATE_LIMIT}.txt
  redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats

done
