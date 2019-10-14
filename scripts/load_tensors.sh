#!/bin/bash

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_load_data)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_load_data not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

DATA_FILE_NAME=${DATA_FILE_NAME:-aibench_generate_data-creditcard-fraud.dat.gz}

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

# Ensure data file is in place
if [ ! -f ${DATA_FILE} ]; then
  echo "Cannot find data file ${DATA_FILE}"
  exit 1
fi

# flush the database
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} flushall

# load the correct AI backend
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} AI.CONFIG LOADBACKEND TF redisai_tensorflow.so

# set the Model
cd $GOPATH/src/github.com/RedisAI/aibench
redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${MODEL_NAME} \
  TF CPU INPUTS transaction reference \
  OUTPUTS output <./tests/models/tensorflow/creditcardfraud.pb

# load the reference data
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
cat ${DATA_FILE} |
  gunzip |
  ${EXE_FILE_NAME} \
    -reporting-period 1000ms \
    -set-blob=false -set-tensor=true \
    -host redis://${DATABASE_HOST}:${DATABASE_PORT} \
    -workers ${NUM_WORKERS} -pipeline 1000 #-max-inserts 10000

redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} info commandstats
