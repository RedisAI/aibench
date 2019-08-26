# AIBench
This repo contains code for benchmarking deep learning solutions,
including RedisAI.
This code is based on a fork of work initially made public by TSBS
at https://github.com/timescale/tsbs.

Current DL solutions supported:

+ RedisAI [(supplemental docs)](docs/redisai.md)
+ TFServing + Redis [(supplemental docs)](docs/tfserving_and_redis.md)
+ Rest API + Redis [(supplemental docs)](docs/restapi_and_redis.md)

## Current use cases


Currently, AIBench supports one use case -- creditcard-fraud from [Kaggle](https://www.kaggle.com/dalpozz/creditcardfraud). This use-case aims to detect a fraudulent transaction. The predictive model to be developed is a neural network implemented in tensorflow with input tensors containing both transaction and reference data.


## Installation

AIBench is a collection of Go programs (with some auxiliary bash and Python
scripts). The easiest way to get and install the Go programs is to use
`go get` and then `go install`:
```bash
# Fetch AIBench and its dependencies
cd $GOPATH/src/github.com/filipecosta90/AIBench/cmd
go get ./...

# Install desired binaries. At a minimum this includes aibench_load_data, and one aibench_run_inference_*
# binary:
cd $GOPATH/src/github.com/filipecosta90/AIBench/cmd
cd aibench_generate_data && go install
cd ../aibench_load_data && go install
cd ../aibench_run_inference_redisai && go install
```

## How to use AIBench

Using AIBench for benchmarking inference performance involves 4 phases: model setup, transaction data parsing and consequent reference data generation, reference data loading, and inference query execution.

### 1. Model setup 
This step is specific for each DL solution being tested ( see Current DL solutions supported above ). 

As an example we will use RedisAI. In that manner, for setting up the model Redis using RedisAI use:
```bash
# flush the database
redis-cli flushall 

# load the correct AI backend
redis-cli AI.CONFIG LOADBACKEND TF redisai_tensorflow/redisai_tensorflow.so

# set the Model
cd $GOPATH/src/github.com/filipecosta90/AIBench
redis-cli -x AI.MODELSET financialNet \
            TF CPU INPUTS transaction reference \
            OUTPUTS output < ./tests/models/tensorflow/creditcardfraud.pb
```

### 2. Transaction data parsing and Reference Data Generation

So that benchmarking results are not affected by generating data on-the-fly, with AIBench you generate the data required for the inference benchmarks first, and then you can (re-)use it as input to the benchmarking phases. 

The following subsection describes in detail the data parsing and data generation required for each specific use case.


#### 2.1 Creditcard-fraud from [Kaggle Competition](https://www.kaggle.com/dalpozz/creditcardfraud). 

The datasets contains transactions made by credit cards in September 2013 by european cardholders. 
The predictive model to be developed is a neural network implemented in tensorflow with input tensors containing both transaction and reference data. 

**Transaction data** contains only numerical input variables which are the result of a PCA transformation, avaible in the following link [csv file](https://www.kaggle.com/mlg-ulb/creditcardfraud#creditcard.csv), resulting into a **1 x 30 input tensor**. 

Likewise, for each **Transaction data**, we generate random deterministic **Reference data**, resulting into numerical input variables to simulated reference data features usually available from cardholders, resulting into an of **1 x 256 input tensor**.

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/AIBench
gunzip -c ./tests/data/creditcard.csv.gz > /tmp/creditcard.csv
aibench_generate_data \
          -input-file /tmp/creditcard.csv \
          -use-case="creditcard-fraud" \
          -seed=12345 \
          -output-file /tmp/aibench_generate_data-creditcard-fraud.bin
```

### 3. Reference Data Loading

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/AIBench 
aibench_load_data \
        -file /tmp/aibench_generate_data-creditcard-fraud.bin \
          -workers 16 
```

### 4. Benchmarking inference performance

To measure inference performance in AIBench, you first need to load
the data using the previous sections. Once the data is loaded,
just use the corresponding `aibench_run_inference_` binary for the database
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/AIBench
cat /tmp/aibench_generate_data-creditcard-fraud.bin \
        | aibench_run_inference_redisai \
       -max-queries 10000 -workers 16 -print-interval 10000 -model financialNet
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