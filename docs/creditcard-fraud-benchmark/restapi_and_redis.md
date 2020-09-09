# aibench Supplemental Guide: DL REST API and Redis


### Benchmarking inference performance -- DL REST API and Redis Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/RedisAI/aibench#reference-data-loading), 
and ensure that the model server is installed and running as explained bellow. If you already have your environment setted up jump to Running DL solution part.

### Load the reference data
Assuming RedisAI is running on port 6379, proceed as follows:
```
# ensure to be at root project folder
cd $GOPATH/src/github.com/RedisAI/aibench

./scripts/load_blobs_redis.sh
```

### Installation -- Install Guinicorn, Flask, and Tensorflow backend on production VM


```bash
cd $GOPATH/src/github.com/RedisAI/aibench/tests/servers/flask

# Install requirements
apt install gunicorn -y
pip install -r requirements.txt
```

#### Test with reference data
```
cd $GOPATH/src/github.com/RedisAI/aibench/tests/servers/flask
export TF_MODEL_PATH=$GOPATH/src/github.com/RedisAI/aibench/tests/models/tensorflow/creditcardfraud.pb
gunicorn --workers=48 --threads 48 -b 0.0.0.0:8000  --log-level error --daemon server:app
```

#### Test without reference data
```
cd $GOPATH/src/github.com/RedisAI/aibench/tests/servers/flask
export TF_MODEL_PATH=$GOPATH/src/github.com/RedisAI/aibench/tests/models/tensorflow/creditcardfraud_noreference.pb
gunicorn --workers=48 --threads 2 -b 0.0.0.0:8000  --log-level error --daemon server_noreference:app
```

---

## Benchmark


Once the data is loaded,
just use the corresponding `aibench_run_inference_flask_tensorflow` binary for the DL Solution
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
./scripts/run_inference_flask_tensorflow.sh
```

### Sample output

```
$ NUM_INFERENCES=10000 ./scripts/run_inference_flask_tensorflow.sh
Benchmarking inference performance with reference data set to: true and model name financialNet
\t\tSaving files with file suffix: flask_tensorflow_ref_redis_true__run_1_workers_1_rate_0.txt
time (ms),total queries,instantaneous inferences/s,overall inferences/s,overall q50 lat(ms),overall q90 lat(ms),overall q95 lat(ms),overall q99 lat(ms),overall q99.999 lat(ms)
159969510478,535,534,534,0.00,0.00,0.00,0.00,0.00
159969510578,1119,584,559,0.00,0.00,0.00,0.00,0.00
159969510678,1722,603,574,0.00,0.00,0.00,0.00,0.00
159969510778,2294,572,573,0.00,0.00,0.00,0.00,0.00
159969510878,2894,600,579,0.00,0.00,0.00,0.00,0.00
159969510978,3498,604,583,0.00,0.00,0.00,0.00,0.00
159969511078,4089,591,584,0.00,0.00,0.00,0.00,0.00
159969511178,4713,624,589,0.00,0.00,0.00,0.00,0.00
159969511278,5331,618,592,0.00,0.00,0.00,0.00,0.00
159969511378,5933,602,593,0.00,0.00,0.00,0.00,0.00
159969511478,6523,590,593,0.00,0.00,0.00,0.00,0.00
159969511578,7131,608,594,0.00,0.00,0.00,0.00,0.00
159969511678,7732,601,595,0.00,0.00,0.00,0.00,0.00
159969511778,8336,604,595,0.00,0.00,0.00,0.00,0.00
159969511878,8944,608,596,0.00,0.00,0.00,0.00,0.00
159969511978,9534,590,596,0.00,0.00,0.00,0.00,0.00
Run complete after 0 inferences with 1 workers (Overall inference rate 596.45 inferences/sec):
All queries:
+ Inference execution latency (statistical histogram):
        min:     0.00 ms,  mean:     0.00 ms, q25:     0.00 ms, med(q50):     0.00 ms, q75:     0.00 ms, q99:     0.00 ms, max:     0.00 ms, stddev:     0.00ms, count: 0, timedOut count: 0

Took:   16.769 sec
Saving Query Latencies HDR Histogram to ~/HIST_flask_tensorflow_ref_redis_true__run_1_workers_1_rate_0.txt
2020/09/10 00:45:20 open ~/HIST_flask_tensorflow_ref_redis_true__run_1_workers_1_rate_0.txt: no such file or directory
Sleeping: 30
```
