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
$ cd dlbench_run_inference_redisai && go install
```

## How to use DLBench

Using DLBench for benchmarking involves 3 phases: data and query
generation, data loading/insertion, and query execution.

### Data and query generation

TBD

#### Data generation

TBD

#### Query generation

TBD

### Benchmarking inference performance

To measure inference performance in DLBench, you first need to load
the data using the previous section and generate the queries as
described earlier. Once the data is loaded and the queries are generated,
just use the corresponding `dlbench_run_inference_` binary for the database
being tested:
```bash
$ dlbench_run_inference_redisai \
       -file /tmp/creditcards.csv \
       -max-queries 10000 -workers 16 -print-interval 1000 
```

You can change the value of the `-workers` flag to
control the level of parallel queries run at the same time. The
resulting output will look similar to this:
```text
(...)


```