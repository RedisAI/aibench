#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_generate_data)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_redisai not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

INPUT_FILE_NAME=${INPUT_FILE_NAME:-./tests/data/creditcard.csv.gz}
TMP_FILE_NAME=${TMP_FILE_NAME:-/tmp/creditcard.csv}

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

cd $GOPATH/src/github.com/RedisAI/aibench
cat ${INPUT_FILE_NAME} |
  gunzip >${TMP_FILE_NAME}
${EXE_FILE_NAME} \
  -input-file ${TMP_FILE_NAME} \
  -use-case="creditcard-fraud" \
  -max-transactions=${NUM_INFERENCES} \
  -seed=${DATA_SEED} >${DATA_FILE}

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Could not generate data file ${DATA_FILE}"
  exit 1
fi

echo "Data generated to file ${DATA_FILE}"
