#!/bin/bash

# flush the database
redis-cli flushall

# load the correct AI backend
redis-cli AI.CONFIG LOADBACKEND TF redisai_tensorflow.so

# set the Model
cd $GOPATH/src/github.com/RedisAI/aibench
redis-cli -x AI.MODELSET financialNet \
            TF CPU INPUTS transaction reference \
            OUTPUTS output < ./tests/models/tensorflow/creditcardfraud.pb

# load the reference data
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz \
        | gunzip \
        | aibench_load_data \
          -reporting-period 1000ms \
          -set-blob=false -set-tensor=true \
          -workers 8 -pipeline 1000


# benchmark inference performance
# make sure you're on the root project folder
redis-cli config resetstat
cd $GOPATH/src/github.com/RedisAI/aibench
cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz \
        | gunzip \
        | aibench_run_inference_redisai \
         -workers 8 \
         -burn-in 10 -max-queries 0 \
         -print-interval 0 -reporting-period 1000ms \
         -model financialNet \
         -host redis://127.0.0.1:6379
redis-cli info commandstats