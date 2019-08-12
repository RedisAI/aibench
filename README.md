# DLBench - DL Benchmark
This repo contains code for benchmarking deep learning solutions,
including RedisAI.
This code is based on a fork of work initially made public by TSBS
at https://github.com/timescale/tsbs.

Current databases supported:

+ RedisAI [(supplemental docs)](docs/redisai.md)

## Overview

TBD 

## Current use cases

TBD


## What the FTSB tests

TBD


## Installation

DLBench is a collection of Go programs (with some auxiliary bash and Python
scripts). The easiest way to get and install the Go programs is to use
`go get` and then `go install`:
```bash
# Fetch DLBench and its dependencies
$ go get github.com/filipecosta90/dlbench
$ cd $GOPATH/src/github.com/filipecosta90/dlbench/cmd
$ go get ./...

# Install desired binaries. At a minimum this includes one dlbench_run_inference_*
# binary:
$ cd $GOPATH/src/github.com/filipecosta90/dlbench/cmd
$ cd dlbench_load_referencedata_redisai && go install
$ cd ../dlbench_run_inference_redisai && go install
```

## How to use DLBench

Using DLBench for benchmarking involves 3 phases: model setup, reference data loading, and inference query execution.

### Model setup



So for setting up the model Redis using RedisAI use:
```bash
# flush the database
$ redis-cli flushall 

# create the index
$ redis-cli AI.CONFIG LOADBACKEND TF redisai_tensorflow/redisai_tensorflow.so
$ redis-cli -x AI.MODELSET financialNet \
            TF CPU INPUTS transaction \
            OUTPUTS classification < ./models/tensorflow/fraudGraph.pb
```


### Reference Data Loading

```bash
$ cat ./data/creditcard.csv.gz \
        | gunzip \
        | dlbench_load_referencedata_redisai \
          -workers 16 
```

### Benchmarking inference performance

To measure inference performance in DLBench, you first need to load
the data using the previous section and generate the queries as
described earlier. Once the data is loaded and the queries are generated,
just use the corresponding `dlbench_run_inference_` binary for the database
being tested:

```bash
$ cat ./data/creditcard.csv.gz \
        | gunzip \
        | dlbench_run_inference_redisai \
       -max-queries 10000 -workers 16 -print-interval 2000 
```

You can change the value of the `-workers` flag to
control the level of parallel queries run at the same time. The
resulting output will look similar to this:

```text
after 2000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.43 ms, q25:     0.33 ms, med(q50):     0.42 ms, q75:     0.49 ms, q99:     0.97 ms, max:     3.71 ms, stddev:     0.30ms, sum: 0.860 sec, count: 2000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.43 ms, q25:     0.33 ms, med(q50):     0.42 ms, q75:     0.49 ms, q99:     0.97 ms, max:     3.71 ms, stddev:     0.30ms, sum: 0.860 sec, count: 2000, timedOut count: 0


after 4000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.41 ms, q25:     0.30 ms, med(q50):     0.40 ms, q75:     0.49 ms, q99:     0.85 ms, max:     3.71 ms, stddev:     0.24ms, sum: 1.633 sec, count: 4000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.41 ms, q25:     0.30 ms, med(q50):     0.40 ms, q75:     0.49 ms, q99:     0.85 ms, max:     3.71 ms, stddev:     0.24ms, sum: 1.633 sec, count: 4000, timedOut count: 0


after 6000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.45 ms, q25:     0.34 ms, med(q50):     0.45 ms, q75:     0.54 ms, q99:     0.83 ms, max:     3.71 ms, stddev:     0.21ms, sum: 2.675 sec, count: 6000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.45 ms, q25:     0.34 ms, med(q50):     0.45 ms, q75:     0.54 ms, q99:     0.83 ms, max:     3.71 ms, stddev:     0.21ms, sum: 2.675 sec, count: 6000, timedOut count: 0


after 8000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.46 ms, q25:     0.37 ms, med(q50):     0.46 ms, q75:     0.54 ms, q99:     0.80 ms, max:     3.71 ms, stddev:     0.19ms, sum: 3.643 sec, count: 8000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.46 ms, q25:     0.37 ms, med(q50):     0.46 ms, q75:     0.54 ms, q99:     0.80 ms, max:     3.71 ms, stddev:     0.19ms, sum: 3.643 sec, count: 8000, timedOut count: 0


++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
Run complete after 10000 queries with 16 workers:
All queries  :
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.46 ms, q25:     0.38 ms, med(q50):     0.47 ms, q75:     0.55 ms, q99:     0.78 ms, max:     3.71 ms, stddev:     0.18ms, sum: 4.639 sec, count: 10000, timedOut count: 0

RedisAI Query:
+ Query execution latency (statistical histogram):
        min:     0.06 ms,  mean:     0.46 ms, q25:     0.38 ms, med(q50):     0.47 ms, q75:     0.55 ms, q99:     0.78 ms, max:     3.71 ms, stddev:     0.18ms, sum: 4.639 sec, count: 10000, timedOut count: 0

Took:    0.317 sec

```