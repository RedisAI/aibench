#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_load_data)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_load_data not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

OUTPUT_NAME_SUFIX=${OUTPUT_NAME_SUFIX:-""}

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Cannot find data file ${DATA_FILE}"
  exit 1
fi

#if [[ "${SETUP_MODEL}" == "true" ]]; then
#
##  # set the Model
cd $GOPATH/src/github.com/RedisAI/aibench
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${MODEL_NAME} \
  TF CPU INPUTS transaction reference \
  OUTPUTS output <./tests/models/tensorflow/creditcardfraud.pb

redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${MODEL_NAME_NOREFERENCE} \
  TF CPU INPUTS transaction \
  OUTPUTS out <./tests/models/tensorflow/creditcardfraud_noreference.pb
#
#fi

# load the reference data
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
cat ${DATA_FILE} |
  ${EXE_FILE_NAME} \
    -reporting-period=1000ms \
    -set-blob=false -set-tensor=true \
    -host=redis://${DATABASE_HOST}:${DATABASE_PORT} \
    -workers=${NUM_WORKERS} -pipeline=${REDIS_PIPELINE_SIZE} 2>&1 | tee ~/redisai_load_tensors_${OUTPUT_NAME_SUFIX}_${NUM_WORKERS}_workers.txt

redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats 2>&1 | tee ~/redisai_load_tensors_commandstats_${OUTPUT_NAME_SUFIX}_${NUM_WORKERS}_workers.txt
