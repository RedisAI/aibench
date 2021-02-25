#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

RED='\033[0;31m'
RESET='\033[0m'

if [[ -z ${BATCHSIZE} ]]; then
  BATCHSIZE=0
fi

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

if [[ "${BACKEND}" = "TF" ]]; then
  MODEL_FILE="./tests/models/tensorflow/mobilenet/mobilenet_v1_100_224_${DEVICE}_NxHxWxC.pb"
  echo "Using Tensorflow backend for the inference."
elif [[ "${BACKEND}" = "TFLITE" ]]; then
  MODEL_FILE="./tests/models/tflite/mobilenet/mobilenet_v1_1.0_224.tflite"
  echo "Using Tensorflow Lite backend for the inference."
  echo -e "${RED}Auto-batching is disabled for Tensorflow Lite${RESET}"
else
  echo "Backend $BACKEND is not supported!"
fi

##  # set the Model
# Auto-batching not supported by the TFLITE backend
if [[ "${BACKEND}" == "TF" ]]; then
    BATCH="BATCHSIZE ${BATCHSIZE}"
fi

resp=$(redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${VISION_MODEL_NAME} \
  ${BACKEND} ${DEVICE} ${BATCH} INPUTS input \
  OUTPUTS MobilenetV1/Predictions/Reshape_1 BLOB <${MODEL_FILE})

if [ "$resp" != "OK" ]; then
    echo "Error loading the model into Redis: ${resp}"
    exit 1
fi
