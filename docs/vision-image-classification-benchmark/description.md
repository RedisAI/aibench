# Image Classification Benchmark 

## Use Case Description 
To assess image classification inference performance, we rely on one network “backbone”: MobileNet V1, which can be considered as one of the standards by the AI community. We’re recurring to COCO 2017 validation dataset (a large-scale object detection, segmentation, and captioning dataset). 

To provide the fairest comparison possible we’ve preprocessed all images by:
- Converting them to tensors of single precision floats
- Normalizing the tensor values to ranges between [0,1]
- Downscaling them so that the smallest dimension ( either Height or Width ) matched 256, and after that downscaled we’ve cropped a random deterministic rectangle of 226x226 as required by the benchmark model. 

All steps are auditable via the following [link](https://github.com/RedisAI/aibench/tree/master/datasets/vision/coco-2017-val). In the following sections you will be provided with the commands required from pre-processing up to benchmarking.



## Installation

The image classification benchmark is a collection of Go programs (with some auxiliary bash and Python
scripts). The easiest way to get and install the Go programs is to use
`go get` and then `go install`, simplified in a make call:
```bash
# Fetch aibench and its dependencies
go get github.com/RedisAI/aibench
cd $GOPATH/src/github.com/RedisAI/aibench
make
```

## How to use aibench's image classification benchmark

Using aibench for benchmarking inference performance involves 3 phases: image preprocessing, model loading, and inference query execution, explained in detail in the following sections.



### 1. Image preprocessing

So that benchmarking results are not affected by pre-processing data on-the-fly, with aibench you pre-process the data required for the inference benchmarks first, and then you can (re-)use it as input to the benchmarking phase. All inference benchmarks use the same dataset, built based uppon the COCO 2017 validation dataset.


```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
make data-vision
```

At the end of the generation you should have a file placed within `/tmp/bulk_data/vision_tensors.out` and the final output as follows:
```
(...)
(...)
Using random seed 12345 to take a random 224 x 224 crop to the scaled image
Saving cropped scaled images to cropped-val2017
100%|█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████| 5000/5000 [00:40<00:00, 124.22it/s]
5000 / 5000 [-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------] 100.00% 202 p/s
Data generated to file /tmp/bulk_data/vision_tensors.out
```

### 2. Model Loading 

As an example of the model loading step we will use RedisAI. You can specificy the `DEVICE=GPU|CPU` in order to load the different device models. In that manner, for setting up the model do as follows:
```bash
cd $GOPATH/src/github.com/RedisAI/aibench
## load the CPU model
$ DEVICE=cpu ./scripts/load_models_mobilenet_redisai.sh

## load the CPU model
$ DEVICE=gpu ./scripts/load_models_mobilenet_redisai.sh
```

### 3. Benchmarking inference performance

To measure inference performance in aibench, you first need to load
the data using the previous sections. Once the data is loaded,
just use the corresponding `aibench_run_inference_` binary for the database
being tested, or use one of the provided scripts to ease the benchmark process.

As an example we will use RedisAI:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench

## run the benchmark
$ ./scripts/run_inference_redisai_mobilenet.sh
```
 
 The
resulting output will look similar to this:

```text
$ ~/go/src/github.com/RedisAI/aibench$ ./scripts/run_inference_redisai_mobilenet.sh
  Benchmarking inference performance with 1 workers. Model name mobilenet_v1_100_224_cpu
  \t\tSaving files with file suffix: redisai__cpu_run_1_workers_1_rate_0.txt
  time (ms),total queries,instantaneous inferences/s,overall inferences/s,overall q50 lat(ms),overall q90 lat(ms),overall q95 lat(ms),overall q99 lat(ms),overall q99.999 lat(ms)
  159674177859,56,56,56,0.00,0.00,0.00,0.00,0.00
  burn-in complete after 100 queries with 1 workers
  159674177959,115,59,57,16.77,17.86,17.86,18.98,18.98
  159674178059,173,58,58,16.93,18.72,19.92,24.48,29.55
  159674178159,230,57,57,16.80,18.72,21.44,29.55,29.84
  159674178259,289,59,58,16.73,18.66,21.15,26.80,29.84
  159674178359,345,56,57,16.83,18.72,20.50,29.55,31.74
  (...)
  159674187059,4612,53,50,18.46,26.69,30.75,40.48,74.05
  159674187159,4667,55,50,18.46,26.57,30.73,40.38,74.05
  159674187259,4723,56,50,18.45,26.46,30.70,40.38,74.05
  159674187359,4778,55,50,18.43,26.41,30.62,40.38,74.05
  159674187459,4833,55,50,18.43,26.34,30.43,40.38,74.05
  159674187559,4887,54,50,18.41,26.27,30.41,39.97,74.05
  159674187659,4938,51,50,18.40,26.22,30.41,40.38,74.05
  159674187759,4991,53,50,18.38,26.19,30.40,40.38,74.05
  Run complete after 4900 inferences with 1 workers (Overall inference rate 49.92 inferences/sec):
  All queries                                      :
  + Inference execution latency (statistical histogram):
          min:    15.51 ms,  mean:    20.07 ms, q25:    16.96 ms, med(q50):    18.38 ms, q75:    21.07 ms, q99:    40.38 ms, max:    74.05 ms, stddev:     5.17ms, count: 4900, timedOut count: 0
  
  RedisAI Query - mobilenet_v1_100_224 :AI.MODELRUN:
  + Inference execution latency (statistical histogram):
          min:    15.51 ms,  mean:    20.07 ms, q25:    16.96 ms, med(q50):    18.38 ms, q75:    21.07 ms, q99:    40.38 ms, max:    74.05 ms, stddev:     5.17ms, count: 4900, timedOut count: 0
  
  Took:  100.162 sec
  Saving Query Latencies HDR Histogram to stats-response-latency-hist.txt
```


### Benchmark variations


You can dive deeper on benchmark configurations by simply recurring to the corresponding binary help, as follows:

```bash
$ aibench_run_inference_redisai_vision --help
Usage of aibench_run_inference_redisai_vision:
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
        model name (default "mobilenet_v1_100_224_cpu_NxHxWxC")
  -output-file-stats-hdr-response-latency-hist string
        File name to output the hdr response latency histogram to (default "stats-response-latency-hist.txt")
  -persist-results
        persist the classification tensors
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