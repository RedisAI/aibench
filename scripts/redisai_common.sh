#!/bin/bash

# Database credentials
DATABASE_HOST=${DATABASE_HOST:-"127.0.0.1"}
DATABASE_PORT=${DATABASE_PORT:-6379}
DATA_SEED=${DATA_SEED:-12345}
MODEL_NAME=${DATABASE_NAME:-"financialNet"}

# Data folder
BULK_DATA_DIR=${BULK_DATA_DIR:-"/tmp/bulk_data"}

# ensure dir exists
mkdir -p ${BULK_DATA_DIR}

# Full path to data file
DATA_FILE=${DATA_FILE:-${BULK_DATA_DIR}/${DATA_FILE_NAME}}

# How many concurrent workers - match num of cores, or default to 4
NUM_WORKERS=${NUM_WORKERS:-$(grep -c ^processor /proc/cpuinfo 2>/dev/null || echo 8)}

set -x
