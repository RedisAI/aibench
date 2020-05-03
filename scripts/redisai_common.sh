#!/bin/bash

# Redis Database
DATABASE_HOST=${DATABASE_HOST:-"127.0.0.1"}
MODELSERVER_HOST=${MODELSERVER_HOST:-"127.0.0.1"}
DATABASE_PORT=${DATABASE_PORT:-6379}

# MySql Database
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-"perf"}
MYSQL_USER=${MYSQL_USER:-"perf"}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-"perf"}
MYSQL_DATABASE=${MYSQL_DATABASE:-"test"}
MYSQL_HOST=${MYSQL_HOST:-"127.0.0.1"}
MYSQL_PORT=${MYSQL_PORT:-3306}

DATA_SEED=${DATA_SEED:-12345}
MODEL_NAME=${MODEL_NAME:-"financialNet"}
MODEL_NAME_NOREFERENCE=${MODEL_NAME_NOREFERENCE:-"financialNet_NoReference"}
SETUP_MODEL=${SETUP_MODEL:-"true"}
REDIS_PIPELINE_SIZE=${REDIS_PIPELINE_SIZE:-100}
DEBUG=${DEBUG:-0}
DATA_FILE_NAME=${DATA_FILE_NAME:-aibench_generate_data-creditcard-fraud.dat}
PRINT_RESPONSES=${PRINT_RESPONSES:-false}
QUERIES_BURN_IN=${QUERIES_BURN_IN:-10000}
SLEEP_BETWEEN_RUNS=${SLEEP_BETWEEN_RUNS:-30}

# Rate limit? if greater than 0 rate is limited.
RATE_LIMIT=${RATE_LIMIT:-0}

# output name
OUTPUT_NAME_SUFIX=${OUTPUT_NAME_SUFIX:-""}

# Data folder
BULK_DATA_DIR=${BULK_DATA_DIR:-"/tmp/bulk_data"}

# ensure dir exists
mkdir -p ${BULK_DATA_DIR}

# Full path to data file
DATA_FILE=${DATA_FILE:-${BULK_DATA_DIR}/${DATA_FILE_NAME}}

# How many concurrent workers - match num of cores, or default to 8
NUM_WORKERS=${NUM_WORKERS:-$(grep -c ^processor /proc/cpuinfo 2>/dev/null || echo 16)}

# How many INFERENCES to Generate/Run
NUM_INFERENCES=${NUM_INFERENCES:-1000000}

set -x
