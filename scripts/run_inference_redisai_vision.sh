#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

RED='\033[0;31m'
RESET='\033[0m'

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

# Ensure data file is in place
if [ ! -f ${OUTPUT_VISION_FILE_NAME} ]; then
  echo "Cannot find data file ${OUTPUT_VISION_FILE_NAME}"
  exit 1
fi

# Auto-batching not supported by the TFLITE backend
if [[ "${BACKEND}" = "TF" ]]; then
  for BATCHSIZE in $(seq ${MIN_BATCHSIZE} ${BATCHSIZE_STEP} ${MAX_BATCHSIZE}); do
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
        hosts=($(echo $DATABASE_HOST | tr "," "\n"))
        ports=($(echo $DATABASE_PORT | tr "," "\n"))
        for i in "${!hosts[@]}"; do
          H="${hosts[i]}"
          P="${ports[i]}"
          redis-cli -h ${H} -p ${P} FLUSHALL
          redis-cli -h ${H} -p ${P} MEMORY PURGE
          redis-cli -h ${H} -p ${P} CONFIG RESETSTAT

          # set the Model
          resp=$(redis-cli -h ${H} -p ${P} -x AI.MODELSET ${VISION_MODEL_NAME} \
            ${BACKEND} ${DEVICE} BATCHSIZE ${BATCHSIZE} INPUTS input \
            OUTPUTS MobilenetV1/Predictions/Reshape_1 BLOB <${MODEL_FILE})

          if [ "$resp" != "OK" ]; then
            echo "Error loading the model into Redis: ${resp}"
            exit 1
          fi

        done

        FILENAME_SUFFIX=redisai_${OUTPUT_NAME_SUFIX}_${DEVICE}_run_${RUN}_workers_${NUM_WORKERS}_autobatching_${BATCHSIZE}_tensorbatchsize_${TENSOR_BATCHSIZE}_rate_${RATE_LIMIT}
        echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
        echo "Benchmarking inference performance with ${NUM_WORKERS} workers."
        echo "   Model name: ${VISION_MODEL_NAME}"
        echo "   Backend: ${BACKEND}"
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
          -reporting-period=1000ms \
          -host=${DATABASE_HOST} \
          -port=${DATABASE_PORT} \
          -json-out-file=./results/JSON_${FILENAME_SUFFIX}.json \
          2>&1 | tee ./results/RAW_${FILENAME_SUFFIX}.txt

        echo "Sleeping: $SLEEP_BETWEEN_RUNS"
        sleep ${SLEEP_BETWEEN_RUNS}
      done
    done
  done

  sleep ${SLEEP_BETWEEN_RUNS}
fi

BATCHSIZE=0

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

      hosts=($(echo $DATABASE_HOST | tr "," "\n"))
      ports=($(echo $DATABASE_PORT | tr "," "\n"))
      for i in "${!hosts[@]}"; do
        H="${hosts[i]}"
        P="${ports[i]}"
        redis-cli -h ${H} -p ${P} FLUSHALL
        redis-cli -h ${H} -p ${P} MEMORY PURGE
        redis-cli -h ${H} -p ${P} CONFIG RESETSTAT

        # set the Model
        if [[ "${BACKEND}" == "TF" ]]; then
            BATCH="BATCHSIZE ${BATCHSIZE}"
        fi
        resp=$(redis-cli -h ${H} -p ${P} -x AI.MODELSET ${VISION_MODEL_NAME} \
          ${BACKEND} ${DEVICE} ${BATCH} INPUTS input \
          OUTPUTS MobilenetV1/Predictions/Reshape_1 BLOB <${MODEL_FILE})

        if [ "$resp" != "OK" ]; then
          echo "Error loading the model into Redis: ${resp}"
          exit 1
        fi

      done

      FILENAME_SUFFIX=redisai_${OUTPUT_NAME_SUFIX}_${DEVICE}_run_${RUN}_workers_${NUM_WORKERS}_autobatching_${BATCHSIZE}_tensorbatchsize_${TENSOR_BATCHSIZE}_rate_${RATE_LIMIT}.txt
      echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
      echo "Benchmarking inference performance with ${NUM_WORKERS} workers."
      echo "   Model name: ${VISION_MODEL_NAME}"
      echo "   Backend: ${BACKEND}"
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
        -reporting-period=1000ms \
        -host=${DATABASE_HOST} \
        -port=${DATABASE_PORT} \
        -json-out-file=./results/JSON_${FILENAME_SUFFIX}.json \
        2>&1 | tee ./results/RAW_${FILENAME_SUFFIX}.txt

      echo "Sleeping: $SLEEP_BETWEEN_RUNS"
      sleep ${SLEEP_BETWEEN_RUNS}
    done
  done
done
