# Credit card fraud benchmark 

## Use Case Description 

We've created a benchmark suite consisting of a fraud-detection use case based on a Kaggle dataset with the extension of reference data. This use case aims to detect a fraudulent transaction based on anonymized credit card transactions and reference data.

We used this benchmark to compare four different AI serving solutions:

- TorchServe: built and maintained by Amazon Web Services (AWS) in collaboration with Facebook, TorchServe is available as part of the PyTorch open-source project.
- Tensorflow Serving: a high-performance serving system, wrapping TensorFlow and maintained by Google.
- Common REST API serving: a common DL production grade setup with Gunicorn (a Python WSGI HTTP server) communicating with Flask through a WSGI protocol, and using TensorFlow as the backend.
-  RedisAI: an AI serving engine for real-time applications built by Redis Labs and Tensorwerk, seamlessly plugged into ​Redis.

We wanted to cover all solutions in an unbiased manner, helping prospective users make an informed decision on the solution that best suits their case, both with and without data locality. 

On top of that, we wanted to reduce the impact of reference data in this benchmark so that it sets the lower bound of what is possible with RedisAI, by:

- Using a high-performance data store common to all solutions (Redis). We wanted to make sure that the bottleneck is not the reference data store and that we stress the model server as much as possible. When the reference data is held in one or more disk-based databases (relational or NoSQL) then RedisAI’s performance benefits will be even greater. 
- Preparing your data in the right format. The reference data for the non-RedisAI solutions is stored in Redis as a blob. This is the best-case scenario. In many applications, data needs to be fetched from different tables and different databases and be prepared into a tensor. RedisAI has a tensor data structure that lets you maintain this reference data in the right format.
- Keeping the reference data to a minimum. The reference data in this benchmark was kept to 1 kilobyte of data but can easily become several megabytes. The larger the reference data, the bigger the impact of data locality on performance. 

## Dataset details 

The initial dataset from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) contains transactions made by credit cards in September 2013 by european cardholders. 
Transaction data contains only numerical input variables which are the result of a PCA transformation, available in the following link [csv file](https://www.kaggle.com/mlg-ulb/creditcardfraud#creditcard.csv), resulting into a numerical value input tensor of size 1 x 30.

We've decided to extend the initial dataset in the sense that for each Transaction data, we generate random deterministic Reference data, commonly used to enrich financial transactions information. In the financial service industry and regulatory agencies, the reference data that defines and describes such financial transactions, can cover all relevant particulars for highly complex transactions with multiple dependencies, entities, and contingencies, thus resulting in a larger numerical value input tensor of size 1 x 256. 


Following the previously described, the predictive model to be developed is a neural network implemented in tensorflow with input tensors containing both transaction (1 x 30 tensor) and reference data (1 x 256 tensor) and with a single output tensor (1 x 2 tensor), presenting the fraudulent and genuine probabilities of each financial transaction.

### Transaction data dataset characteristics

The creditcard-fraud dataset from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) contains transactions that occurred in two days, where we have 492 frauds out of 284,807 transactions. Each transaction data tensor represents 120 Bytes of Data ( 30 x 4 Bytes ), whereas each reference data tensor represents 1024 Bytes of Data ( 256 * 4 Bytes ). 

## Installation

The credit card fraud benchmark is a collection of Go programs (with some auxiliary bash and Python
scripts). The easiest way to get and install the Go programs is to use
`go get` and then `go install`, simplified in a make call:
```bash
# Fetch aibench and its dependencies
go get github.com/RedisAI/aibench
cd $GOPATH/src/github.com/RedisAI/aibench
git checkout v0.1.1
make
```

## How to use aibench's Credit card fraud benchmark

Using aibench for benchmarking inference performance involves 3 phases: transaction data parsing and consequent reference data generation, reference data and model loading, and inference query execution, explained in detail in the following sections.


### 1. Transaction data parsing and Reference Data Generation

So that benchmarking results are not affected by generating data on-the-fly, with aibench you generate the data required for the inference benchmarks first, and then you can (re-)use it as input to the benchmarking and reference data loading phases. All inference benchamarks use the same dataset, built based uppon the [Kaggle Financial Transactions Dataset csv file](https://www.kaggle.com/mlg-ulb/creditcardfraud#creditcard.csv) and random deterministic Reference data, if using the same random seed.


```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
make data
```

### 1. Model Loading and Reference Data Loading

We consider that the reference data that defines and describes the financial transactions already resides on a datastore common to all benchmarks. We've decided to use Redis as the primary (and only) datastore for the inference benchmarks. The reference data tensors will be stored in redis in two distinct formats:
 - RedisAI BLOB Tensor, following the pattern referenceTensor:[uniqueId]. 
 - Redis Binary-safe strings, following the pattern referenceBLOB:[uniqueId].
 
Depending on the DL solution being tested the benchmark will use one of the above ( see Current DL solutions supported above and followed the detail links ). 

As an example we will use RedisAI. In that manner, for setting up the model and loading the reference data do as follows:
```bash
cd $GOPATH/src/github.com/RedisAI/aibench
## load the reference tensors ( this will also modelset, etc... )
$ ./scripts/load_tensors_redis.sh
```

### 4. Benchmarking inference performance

To measure inference performance in aibench, you first need to load
the data using the previous sections. Once the data is loaded,
just use the corresponding `aibench_run_inference_` binary for the database
being tested, or use one of the provided scripts to ease the benchmark process.

As an example we will use RedisAI:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench

## run the benchmark
$ ./scripts/run_inference_redisai.sh
```
 
 The
resulting output will look similar to this:

```text
$ cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz         | gunzip         | aibench_run_inference_redisai          -workers 16          -burn-in 10 -max-queries 100010          -print-interval 10000 -reporting-period 0ms          -model financialNet          -host redis://#########:6379 

burn-in complete after 10 queries with 16 workers
after 9990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.59 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.97 ms, q75:     1.01 ms, q99:     1.30 ms, max:     4.36 ms, stddev:     0.15ms, sum: 9.739 sec, count: 9990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.59 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.97 ms, q75:     1.01 ms, q99:     1.30 ms, max:     4.36 ms, stddev:     0.15ms, sum: 9.739 sec, count: 9990, timedOut count: 0


after 19990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.21 ms, max:     4.36 ms, stddev:     0.12ms, sum: 19.363 sec, count: 19990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.21 ms, max:     4.36 ms, stddev:     0.12ms, sum: 19.363 sec, count: 19990, timedOut count: 0


after 29990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.16 ms, max:     4.36 ms, stddev:     0.10ms, sum: 28.971 sec, count: 29990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.16 ms, max:     4.36 ms, stddev:     0.10ms, sum: 28.971 sec, count: 29990, timedOut count: 0


after 39990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.14 ms, max:     4.36 ms, stddev:     0.09ms, sum: 38.562 sec, count: 39990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.14 ms, max:     4.36 ms, stddev:     0.09ms, sum: 38.562 sec, count: 39990, timedOut count: 0


after 49990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.13 ms, max:     4.36 ms, stddev:     0.09ms, sum: 48.114 sec, count: 49990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.13 ms, max:     4.36 ms, stddev:     0.09ms, sum: 48.114 sec, count: 49990, timedOut count: 0


after 59990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 57.832 sec, count: 59990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 57.832 sec, count: 59990, timedOut count: 0


after 69990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.97 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 67.613 sec, count: 69990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.97 ms, q25:     0.92 ms, med(q50):     0.97 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 67.613 sec, count: 69990, timedOut count: 0


after 79990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 77.020 sec, count: 79990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 77.020 sec, count: 79990, timedOut count: 0


after 89990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 86.538 sec, count: 89990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.48 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 86.538 sec, count: 89990, timedOut count: 0


after 99990 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.38 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 96.015 sec, count: 99990, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.38 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 96.015 sec, count: 99990, timedOut count: 0


++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
Run complete after 100000 queries with 16 workers:
All queries                                                 :
+ Inference execution latency (statistical histogram):
	min:     0.38 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 96.024 sec, count: 100000, timedOut count: 0

RedisAI Query - with AI.TENSORSET transacation datatype BLOB:
+ Inference execution latency (statistical histogram):
	min:     0.38 ms,  mean:     0.96 ms, q25:     0.92 ms, med(q50):     0.96 ms, q75:     1.00 ms, q99:     1.12 ms, max:     4.36 ms, stddev:     0.08ms, sum: 96.024 sec, count: 100000, timedOut count: 0

Took:    6.048 sec
```


### Benchmark variations


You can dive deeper on benchmark configurations by simply recurring to the corresponding binary help, as follows:

```bash
$ aibench_run_inference_redisai --help
Usage of aibench_run_inference_redisai:
  -burn-in uint
        Number of queries to ignore before collecting statistics.
  -cluster-mode
        read cluster slots and distribute inferences among shards.
  -cpuprofile string
        Write a cpu profile to this file.
  -debug int
        Whether to print debug messages.
  -enable-reference-data-mysql
        Whether to enable benchmarking inference with a model with reference data on MySql or not (default false).
  -enable-reference-data-redis
        Whether to enable benchmarking inference with a model with reference data on Redis or not (default false).
  -file string
        File name to read queries from
  -host string
        Redis host address, if more than one is passed will round robin requests (default "localhost")
  -ignore-errors
        Whether to ignore the inference errors and continue. By default on error the benchmark stops (default false).
  -limit-rps uint
        Limit overall RPS. 0 disables limit.
  -max-queries uint
        Limit the number of queries to send, 0 = no limit
  -memprofile string
        Write a memory profile to this file.
  -model string
        model name
  -model-filename string
        modelFilename
  -output-file-stats-hdr-response-latency-hist string
        File name to output the hdr response latency histogram to (default "stats-response-latency-hist.txt")
  -pool-pipeline-concurrency int
        If limit is zero then no limit will be used and pipelines will only be limited by the specified time window
  -pool-pipeline-window duration
        If window is zero then implicit pipelining will be disabled (default 500µs)
  -port string
        Redis host port, if more than one is passed will round robin requests (default "6379")
  -prewarm-queries
        Run each inference twice in a row so the warm inference is guaranteed to be a cache hit
  -print-interval uint
        Print timing stats to stderr after this many queries (0 to disable) (default 1000)
  -print-responses
        Pretty print response bodies for correctness checking (default false).
  -repetitions uint
        Number of repetitions of requests per dataset ( will round robin ). (default 10)
  -reporting-period duration
        Period to report write stats (default 1s)
  -seed int
        PRNG seed (default, or 0, uses the current timestamp).
  -use-dag
        use DAGRUN
  -workers uint
        Number of concurrent requests to make. (default 8)

```
