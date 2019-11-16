#!/bin/bash

# Install desired binaries. At a minimum this includes aibench_generate_data, aibench_load_data, and one aibench_run_inference_*
# binary:
cd $GOPATH/src/github.com/RedisAI/aibench/cmd
cd aibench_generate_data && go build && go install
cd ../aibench_load_data && go build && go install
cd ../aibench_run_inference_redisai && go build && go install
cd ../aibench_run_inference_tensorflow_serving && go build && go install
cd ../aibench_run_inference_flask_tensorflow && go build && go install
