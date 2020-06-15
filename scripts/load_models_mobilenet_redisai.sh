#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

##  # set the Model
cd $GOPATH/src/github.com/RedisAI/aibench
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${VISION_MODEL_NAME} \
  TF ${DEVICE} INPUTS input \
  OUTPUTS MobilenetV1/Predictions/Reshape_1 BLOB <./tests/models/tensorflow/mobilenet/mobilenet_v1_100_224_${DEVICE}_NxHxWxC.pb
