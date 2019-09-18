# aibench
This repo contains code for benchmarking deep learning solutions,
including RedisAI.
This code is based on a fork of work initially made public by TSBS
at https://github.com/timescale/tsbs.

Current DL solutions supported:

+ RedisAI [(supplemental docs)](docs/redisai.md)
+ TFServing + Redis [(supplemental docs)](docs/tfserving_and_redis.md)
+ Rest API + Redis [(supplemental docs)](docs/restapi_and_redis.md)

## Current use cases


Currently, aibench supports one use case -- creditcard-fraud from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) with the extension of reference data. This use-case aims to detect a fraudulent transaction based on 
anonymized credit card transactions and reference data. 

The initial dataset from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) contains transactions made by credit cards in September 2013 by european cardholders. 
Transaction data contains only numerical input variables which are the result of a PCA transformation, available in the following link [csv file](https://www.kaggle.com/mlg-ulb/creditcardfraud#creditcard.csv), resulting into a numerical value input tensor of size 1 x 30.

We've decided to extend the initial dataset in the sense that for each Transaction data, we generate random deterministic Reference data, commonly used to enrich financial transactions information. In the financial service industry and regulatory agencies, the reference data that defines and describes such financial transactions, can cover all relevant particulars for highly complex transactions with multiple dependencies, entities, and contingencies, thus resulting in a larger numerical value input tensor of size 1 x 256. 


Following the previously described, the predictive model to be developed is a neural network implemented in tensorflow with input tensors containing both transaction (1 x 30 tensor) and reference data (1 x 256 tensor) and with a single output tensor (1 x 2 tensor), presenting the fraudulent and genuine probabilities of each financial transaction.

### Transaction data dataset characteristics

The creditcard-fraud dataset from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) contains transactions that occurred in two days, where we have 492 frauds out of 284,807 transactions. Each transaction data tensor represents 120 Bytes of Data ( 30 x 4 Bytes ), whereas each reference data tensor represents 1024 Bytes of Data ( 256 * 4 Bytes ). 
## Installation

aibench is a collection of Go programs (with some auxiliary bash and Python
scripts). The easiest way to get and install the Go programs is to use
`go get` and then `go install`:
```bash
# Fetch aibench and its dependencies
cd $GOPATH/src/github.com/RedisAI/aibench/cmd
go get ./...

# Install desired binaries. At a minimum this includes aibench_load_data, and one aibench_run_inference_*
# binary:
cd $GOPATH/src/github.com/RedisAI/aibench/cmd
cd aibench_generate_data && go install
cd ../aibench_load_data && go install
cd ../aibench_run_inference_redisai && go install
cd ../aibench_run_inference_tensorflow_serving && go install
cd ../aibench_run_inference_flask_tensorflow && go install
```

## How to use aibench

Using aibench for benchmarking inference performance involves 4 phases: model setup, transaction data parsing and consequent reference data generation, reference data loading, and inference query execution, explained in detail in the following sections.

### 1. Model setup 
This step is specific for each DL solution being tested ( see Current DL solutions supported above ). 

As an example we will use RedisAI. In that manner, for setting up the model Redis using RedisAI use:
```bash
# flush the database
redis-cli flushall 

# load the correct AI backend
redis-cli AI.CONFIG LOADBACKEND TF redisai_tensorflow.so

# set the Model
cd $GOPATH/src/github.com/RedisAI/aibench
redis-cli -x AI.MODELSET financialNet \
            TF CPU INPUTS transaction reference \
            OUTPUTS output < ./tests/models/tensorflow/creditcardfraud.pb
```

### 2. Transaction data parsing and Reference Data Generation

So that benchmarking results are not affected by generating data on-the-fly, with aibench you generate the data required for the inference benchmarks first, and then you can (re-)use it as input to the benchmarking and reference data loading phases. All inference benchamarks use the same dataset, built based uppon the [Kaggle Financial Transactions Dataset csv file](https://www.kaggle.com/mlg-ulb/creditcardfraud#creditcard.csv) and random deterministic Reference data, if using the same random seed.


```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
cat ./tests/data/creditcard.csv.gz \
          | gunzip > /tmp/creditcard.csv
aibench_generate_data \
          -input-file /tmp/creditcard.csv \
          -use-case="creditcard-fraud" \
          -seed=12345 \
          | gzip > /tmp/aibench_generate_data-creditcard-fraud.dat.gz
```

### 3. Reference Data Loading

We consider that the reference data that defines and describes the financial transactions already resides on a datastore common to all benchmarks. We've decided to use Redis as the primary (and only) datastore for the inference benchmarks. The reference data tensors will be stored in redis in two distinct formats:
 - RedisAI BLOB Tensor, following the pattern referenceTensor:[uniqueId]. 
 - Redis Binary-safe strings, following the pattern referenceBLOB:[uniqueId].
 
 An an example for the referenceData 1 aibench_load_data will issue the following commands:

 ```
"AI.TENSORSET" "referenceTensor:1" "FLOAT" "1" "256" "BLOB" "( binary data representation of [256]float32 )"
"SET" "referenceBLOB:1" "( binary data representation of [256]float32 )"
 ```

To fully contain the dataset, the datastore will require at minimum 570MB of space, which already accounts for keys used memory space. 

 After having executed step 2 ( aibench_generate_data ), you can proceed with the reference data loading to the primary datastore, issuing the following command:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench 
cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz \
        | gunzip \
        | aibench_load_data \
          -reporting-period 1000ms \
          -set-blob=false -set-tensor=true \
          -workers 16 -pipeline 100
```

### 4. Benchmarking inference performance

To measure inference performance in aibench, you first need to load
the data using the previous sections. Once the data is loaded,
just use the corresponding `aibench_run_inference_` binary for the database
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz \
        | gunzip \
        | aibench_run_inference_redisai \
         -workers 8 \
         -burn-in 10 -max-queries 100010 \
         -print-interval 0 -reporting-period 1000ms \
         -model financialNet \
         -host redis://127.0.0.1:6379
redis-cli info commandstats
```

You can change the value of the `-workers` flag to
control the level of parallel queries run at the same time. The
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