#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

# Ensure generator is available
EXE_FILE_NAME=${EXE_FILE_NAME:-$(which aibench_run_inference_redisai_vision)}
if [[ -z "${EXE_FILE_NAME}" ]]; then
  echo "aibench_run_inference_redisai_vision not available. It is not specified explicitly and not found in \$PATH"
  exit 1
fi

# create results dir if doesnt exist
mkdir -p ./results

# Load parameters - common
EXE_DIR=${EXE_DIR:-$(dirname $0)}
source ${EXE_DIR}/redisai_common.sh

# Ensure data file is in place
if [ ! -f ${OUTPUT_VISION_FILE_NAME} ]; then
  echo "Cannot find data file ${OUTPUT_VISION_FILE_NAME}"
  exit 1
fi

for BATCHSIZE in $(seq ${MIN_BATCHSIZE} ${BATCHSIZE_STEP} ${MAX_BATCHSIZE}); do
  if [ $BATCHSIZE == 0 ]; then
    BATCHSIZE=1
  fi
  echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
  echo "@@@@@@@@@@@@@@@@@@ AUTO-BATCHING ${BATCHSIZE} @@@@@@@@@@@@@@@@@@"
  echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"

  # we overload the NUM_WORKERS here for the official benchmark
  for NUM_WORKERS in $(seq ${MIN_CLIENTS} ${CLIENTS_STEP} ${MAX_CLIENTS}); do
    if [ $NUM_WORKERS == 0 ]; then
      NUM_WORKERS=1
    fi
    if [[ "$NUM_WORKERS" -lt "$BATCHSIZE" ]]; then
      echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
      echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
      echo "   Skipping this loop due to "$NUM_WORKERS" -lt "$BATCHSIZE""
      echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
      echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
      continue
    fi
    TENSOR_BATCHSIZE=1

    for RUN in $(seq 1 ${RUNS_PER_VARIATION}); do

      # flushall
      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} FLUSHALL

      # set the Model
      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${VISION_MODEL_NAME} \
        TF ${DEVICE} BATCHSIZE ${BATCHSIZE} INPUTS input \
        OUTPUTS MobilenetV1/Predictions/Reshape_1 BLOB <./tests/models/tensorflow/mobilenet/mobilenet_v1_100_224_${DEVICE}_NxHxWxC.pb

      FILENAME_SUFFIX=redisai_${OUTPUT_NAME_SUFIX}_${DEVICE}_run_${RUN}_workers_${NUM_WORKERS}_autobatching_${BATCHSIZE}_tensorbatchsize_${TENSOR_BATCHSIZE}_rate_${RATE_LIMIT}
      echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
      echo "Benchmarking inference performance with ${NUM_WORKERS} workers."
      echo "   Model name: ${VISION_MODEL_NAME}"
      echo "   Server side autobatching size: ${BATCHSIZE}"
      echo "   Tensor batch size: ${TENSOR_BATCHSIZE}"
      echo "   Saving files with file suffix: ${FILENAME_SUFFIX}"
      echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
      # benchmark inference performance
      # make sure you're on the root project folder

      ${EXE_FILE_NAME} \
        --file=${OUTPUT_VISION_FILE_NAME} \
        -model=${VISION_MODEL_NAME} \
        -debug=${DEBUG} \
        -workers=${NUM_WORKERS} \
        -metadata-autobatching=${BATCHSIZE} \
        -batch-size=${TENSOR_BATCHSIZE} \
        -burn-in=${VISION_QUERIES_BURN_IN} -max-queries=${NUM_VISION_INFERENCES} \
        -print-interval=0 -reporting-period=1000ms \
        -host=${DATABASE_HOST} \
        -port=${DATABASE_PORT} \
        -json-out-file=./results/JSON_${FILENAME_SUFFIX}.json \
        2>&1 | tee ./results/RAW_${FILENAME_SUFFIX}.txt

      echo "Sleeping: $SLEEP_BETWEEN_RUNS"
      sleep ${SLEEP_BETWEEN_RUNS}
    done
  done

done

BATCHSIZE=1

# we overload the NUM_WORKERS here for the official benchmark
for NUM_WORKERS in $(seq ${MIN_CLIENTS} ${CLIENTS_STEP} ${MAX_CLIENTS}); do
  if [ $NUM_WORKERS == 0 ]; then
    NUM_WORKERS=1
  fi

  # we overload the NUM_WORKERS here for the official benchmark
  for TENSOR_BATCHSIZE in $(seq ${MIN_TENSOR_BATCHSIZE} ${TENSOR_BATCHSIZE_STEP} ${MAX_TENSOR_BATCHSIZE}); do
    if [ $TENSOR_BATCHSIZE == 0 ]; then
      TENSOR_BATCHSIZE=1
    fi
    echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
    echo "@@@@@@@@@@@@@@@@@@ TENSOR-BATCHING ${TENSOR_BATCHSIZE} @@@@@@@@@@@@@@@@@@"
    echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"

    for RUN in $(seq 1 ${RUNS_PER_VARIATION}); do

      # flushall
      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} FLUSHALL

      # set the Model
      redis-cli -h ${DATABASE_HOST} -p ${DATABASE_PORT} -x AI.MODELSET ${VISION_MODEL_NAME} \
        TF ${DEVICE} BATCHSIZE ${BATCHSIZE} INPUTS input \
        OUTPUTS MobilenetV1/Predictions/Reshape_1 BLOB <./tests/models/tensorflow/mobilenet/mobilenet_v1_100_224_${DEVICE}_NxHxWxC.pb

      FILENAME_SUFFIX=redisai_${OUTPUT_NAME_SUFIX}_${DEVICE}_run_${RUN}_workers_${NUM_WORKERS}_autobatching_${BATCHSIZE}_tensorbatchsize_${TENSOR_BATCHSIZE}_rate_${RATE_LIMIT}.txt
      echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
      echo "Benchmarking inference performance with ${NUM_WORKERS} workers."
      echo "   Model name: ${VISION_MODEL_NAME}"
      echo "   Server side autobatching size: ${BATCHSIZE}"
      echo "   Tensor batch size: ${TENSOR_BATCHSIZE}"
      echo "   Saving files with file suffix: ${FILENAME_SUFFIX}"
      echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
      # benchmark inference performance
      # make sure you're on the root project folder

      ${EXE_FILE_NAME} \
        --file=${OUTPUT_VISION_FILE_NAME} \
        -model=${VISION_MODEL_NAME} \
        -debug=${DEBUG} \
        -workers=${NUM_WORKERS} \
        -metadata-autobatching=${BATCHSIZE} \
        -batch-size=${TENSOR_BATCHSIZE} \
        -burn-in=${VISION_QUERIES_BURN_IN} -max-queries=${NUM_VISION_INFERENCES} \
        -print-interval=0 -reporting-period=1000ms \
        -host=${DATABASE_HOST} \
        -port=${DATABASE_PORT} \
        -json-out-file=./results/JSON_${FILENAME_SUFFIX}.json \
        2>&1 | tee ./results/RAW_${FILENAME_SUFFIX}.txt

      echo "Sleeping: $SLEEP_BETWEEN_RUNS"
      sleep ${SLEEP_BETWEEN_RUNS}
    done
  done
done
