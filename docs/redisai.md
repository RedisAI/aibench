# aibench Supplemental Guide: RedisAI


### Benchmarking inference performance -- RedisAI Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/filipecosta90/aibench#reference-data-loading). 

Once the data is loaded,
just use the corresponding `aibench_run_inference_redisai` binary for the DL Solution
being tested:

```bash
# make sure you're on the root project folder
$ cd $GOPATH/src/github.com/filipecosta90/aibench
$ cat ./tests/data/creditcard.csv.gz \
        | gunzip \
        | aibench_run_inference_redisai \
         -max-queries 10000 -workers 16 -print-interval 2000 \
         -model financialNet \
         -redis-host 127.0.0.1:6379 
```

#### Sequence diagram - RedisAI Solution

The following diagram illustrates the sequence of requests made for each inference.


![Sequence diagram - RedisAI Solution][aibench_client_redisai]

[aibench_client_redisai]: ./aibench_client_redisai.png