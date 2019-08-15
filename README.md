# DLBench - DL Benchmark
This repo contains code for benchmarking deep learning solutions,
including RedisAI.
This code is based on a fork of work initially made public by TSBS
at https://github.com/timescale/tsbs.

Current DL solutions supported:

+ RedisAI [(supplemental docs)](docs/redisai.md)
+ TFServing + Redis [(supplemental docs)](docs/tfserving_and_redis.md)
+ Rest API + Redis [(supplemental docs)](docs/restapi_and_redis.md)

## Current use cases


Currently, DLBench supports one use case -- creditcard-fraud from [Kaggle](https://www.kaggle.com/dalpozz/creditcardfraud). This use-case aims to detect a fraudulent transaction. The predictive model to be developed is a neural network implemented in tensorflow with input tensors containing both transaction and reference data.


## Installation

DLBench is a collection of Go programs (with some auxiliary bash and Python
scripts). The easiest way to get and install the Go programs is to use
`go get` and then `go install`:
```bash
# Fetch DLBench and its dependencies
go get github.com/filipecosta90/dlbench
cd $GOPATH/src/github.com/filipecosta90/dlbench/cmd
go get ./...

# Install desired binaries. At a minimum this includes dlbench_load_referencedata, and one dlbench_run_inference_*
# binary:
cd $GOPATH/src/github.com/filipecosta90/dlbench/cmd
cd dlbench_load_referencedata && go install
cd ../dlbench_run_inference_redisai && go install
```

## How to use DLBench

Using DLBench for benchmarking involves 3 phases: model setup, reference data loading, and inference query execution.

### Model setup



So for setting up the model Redis using RedisAI use:
```bash
# flush the database
redis-cli flushall 

# load the correct AI backend
redis-cli AI.CONFIG LOADBACKEND TF redisai_tensorflow/redisai_tensorflow.so

# set the Model
redis-cli -x AI.MODELSET financialNet \
            TF CPU INPUTS transaction reference \
            OUTPUTS output < ./models/tensorflow/creditcardfraud.pb
```

### Reference Data Loading

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/dlbench
cat ./data/creditcard.csv.gz \
        | gunzip \
        | dlbench_load_referencedata \
          -workers 16 
```

### Benchmarking inference performance

To measure inference performance in DLBench, you first need to load
the data using the previous section and generate the queries as
described earlier. Once the data is loaded and the queries are generated,
just use the corresponding `dlbench_run_inference_` binary for the database
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/dlbench
cat ./data/creditcard.csv.gz \
        | gunzip \
        | dlbench_run_inference_redisai \
       -max-queries 10000 -workers 16 -print-interval 2000 -model financialNet
```

You can change the value of the `-workers` flag to
control the level of parallel queries run at the same time. The
resulting output will look similar to this:

```text
after 2000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    22.70 ms, q25:    22.37 ms, med(q50):    22.71 ms, q75:    23.01 ms, q99:    24.09 ms, max:    34.71 ms, stddev:     0.77ms, sum: 45.391 sec, count: 2000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    22.70 ms, q25:    22.37 ms, med(q50):    22.71 ms, q75:    23.01 ms, q99:    24.09 ms, max:    34.71 ms, stddev:     0.77ms, sum: 45.391 sec, count: 2000, timedOut count: 0


after 4000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    23.34 ms, q25:    22.36 ms, med(q50):    22.71 ms, q75:    23.09 ms, q99:    32.35 ms, max:    34.71 ms, stddev:     2.44ms, sum: 93.367 sec, count: 4000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    23.34 ms, q25:    22.36 ms, med(q50):    22.71 ms, q75:    23.09 ms, q99:    32.35 ms, max:    34.71 ms, stddev:     2.44ms, sum: 93.367 sec, count: 4000, timedOut count: 0


after 6000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    24.22 ms, q25:    22.36 ms, med(q50):    22.80 ms, q75:    23.94 ms, q99:    32.27 ms, max:    34.71 ms, stddev:     3.17ms, sum: 145.314 sec, count: 6000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    24.22 ms, q25:    22.36 ms, med(q50):    22.80 ms, q75:    23.94 ms, q99:    32.27 ms, max:    34.71 ms, stddev:     3.17ms, sum: 145.314 sec, count: 6000, timedOut count: 0


after 8000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    24.37 ms, q25:    22.44 ms, med(q50):    22.97 ms, q75:    26.48 ms, q99:    32.10 ms, max:    34.71 ms, stddev:     3.01ms, sum: 194.988 sec, count: 8000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    24.37 ms, q25:    22.44 ms, med(q50):    22.97 ms, q75:    26.48 ms, q99:    32.10 ms, max:    34.71 ms, stddev:     3.01ms, sum: 194.988 sec, count: 8000, timedOut count: 0


++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
Run complete after 10000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    24.21 ms, q25:    22.50 ms, med(q50):    23.05 ms, q75:    25.40 ms, q99:    31.84 ms, max:    34.71 ms, stddev:     2.77ms, sum: 242.109 sec, count: 10000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:    16.18 ms,  mean:    24.21 ms, q25:    22.50 ms, med(q50):    23.05 ms, q75:    25.40 ms, q99:    31.84 ms, max:    34.71 ms, stddev:     2.77ms, sum: 242.109 sec, count: 10000, timedOut count: 0

Took:   15.151 sec
```