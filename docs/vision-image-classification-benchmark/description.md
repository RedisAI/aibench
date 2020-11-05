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

## load the GPU model
$ DEVICE=gpu ./scripts/load_models_mobilenet_redisai.sh
```

#### 2.1 Auto batching
By default, the benchmark uses a batch size of 0. You can benchmark RedisAI auto batching capabilities by specifying the `BATCHSIZE=<n>` env variable.

When provided with an n that is greater than 0, the engine will batch incoming requests from multiple clients that use the model with input tensors of the same shape. 

Please denote that single client benchmarks will not benefit from auto-batching. 

In that manner, for setting up the model with auto batching up to 32 tensors from distinct clients, do as follows:

```bash
cd $GOPATH/src/github.com/RedisAI/aibench
## load the CPU model and specify auto batching up to 32 incoming tensors from distinct clients
$ DEVICE=cpu BATCHSIZE=32 ./scripts/load_models_mobilenet_redisai.sh

## load the GPU model and specify auto batching up to 32 incoming tensors from distinct clients
$ DEVICE=gpu BATCHSIZE=32 ./scripts/load_models_mobilenet_redisai.sh
```

### 3. Benchmarking inference performance

To measure inference performance in aibench, you first need to load
the data using the previous sections. Once the data is loaded,
just use the corresponding `aibench_run_inference_` binary for the model server
being tested, or use one of the provided scripts to ease the benchmark process.

As an example we will use RedisAI:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench

## run the benchmark
$ ./scripts/run_inference_redisai_vision.sh
```
 
 The
resulting output will look similar to this:

```text
$ ~/go/src/github.com/RedisAI/aibench$ ./scripts/run_inference_redisai_vision.sh
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

#### 3.1 Batching multiple inputs (images) to 4D batch tensor

The used model is written to produce outputs from a batch of multiple inputs at the same time, 
with input tensor having a B x C x H x W layout. 

By default, the benchmark uses a 1 x C x H x W layout, meaning that each input tensor represents a single image. 

Please denote that Batching multiple inputs (images) to 4D batch tensor is done on client side and completly independent of auto-batching settings on the server. Single client benchmarks can benefit from this benchmark feature. 

In that manner, for batching 10 images into a single tensor and run a single modelrun, do as follows:

```bash
cd $GOPATH/src/github.com/RedisAI/aibench
## Run the benchmark with the CPU model and batch 10 inputs (images) to 4D batch tensor
$ DEVICE=cpu TENSOR_BATCHSIZE=10 ./scripts/run_inference_redisai_vision.sh

## Run the benchmark with the GPU model and batch 10 inputs (images) to 4D batch tensor
$ DEVICE=gpu TENSOR_BATCHSIZE=10 ./scripts/run_inference_redisai_vision.sh
```

### 4. Retrieving additional AI Module/Models runtime stats

You can retrieve additional runtime stats by leveraging the following 3 commands:

- `AI.INFO <model key>` -- to retrieve statistics like the cumulative duration of executions in microseconds, total number of executions and average batch size ( by dividing SAMPLES per CALLS ). Full details on the [following link](https://oss.redislabs.com/redisai/commands/#aiinfo) 

For the given example of batching 10 images per modelrun, AI.INFO reply should look like the following:
```
$ redis-cli AI.INFO mobilenet_v1_100_224_cpu 
 1) "key"
 2) "mobilenet_v1_100_224_cpu"
 3) "type"
 4) "MODEL"
 5) "backend"
 6) "TF"
 7) "device"
 8) "cpu"
 9) "tag"
10) ""
11) "duration"
12) (integer) 76611636
13) "samples"
14) (integer) 5000
15) "calls"
16) (integer) 500
17) "errors"
18) (integer) 0
```

- `INFO COMMANDSTATS` -- To retrieve the cumulative main thread execution time of the commands. 

For the given example of batching 10 images per modelrun, `INFO COMMANDSTATS` reply should look like the following:

```
$ redis-cli info commandstats
# Commandstats
cmdstat_ai.modelrun:calls=500,usec=10383,usec_per_call=20.77
cmdstat_ai.tensorget:calls=500,usec=3148,usec_per_call=6.30
cmdstat_ai.tensorset:calls=500,usec=1929229,usec_per_call=3858.46
```


- `INFO MODULES` -- To retrieve per device CPU usage stats as well as some important load time configs.

For the given example of batching 10 images per modelrun, `INFO MODULES` reply should look like the following:
```
$ redis-cli info modules
# Modules
module:name=ai,ver=999999,api=1,filters=0,usedby=[],using=[],options=[]

# ai_git
ai_git_sha:deb65404af7d500dd257bdafc231815fee82e5f8

# ai_load_time_configs
ai_threads_per_queue:1
ai_inter_op_parallelism:0
ai_intra_op_parallelism:0

# ai_cpu
ai_self_used_cpu_sys:133.079150
ai_self_used_cpu_user:1459.490824
ai_children_used_cpu_sys:0.001464
ai_children_used_cpu_user:0.001836
ai_queue_CPU_bthread_#1_used_cpu_total:0.000359
```

