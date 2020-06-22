#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_generate_data_vision)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_redisai not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi


# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

cd $GOPATH/src/github.com/RedisAI/aibench

${EXE_FILE_NAME} \
  --input-val-dir=${INPUT_VISION_VAL_DIR} \
  --output-file=${OUTPUT_VISION_FILE_NAME}

# Ensure data file is in place
if [ ! -f ${OUTPUT_VISION_FILE_NAME} ]; then
  echo "Could not generate data file ${OUTPUT_VISION_FILE_NAME}"
  exit 1
fi

echo "Data generated to file ${OUTPUT_VISION_FILE_NAME}"
