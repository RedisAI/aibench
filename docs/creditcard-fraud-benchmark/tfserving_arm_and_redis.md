# aibench Supplemental Guide: TensorFlow Serving on ARM and Redis

This page documents the ARM variation for the fraud-detection use case using TensorFlow Serving on ARM and Redis. 
To do so, we rely on "TensorFlow Serving ARM project" docker images, specifically the 2.3.0 version, as documented [here](https://github.com/emacski/tensorflow-serving-arm/tree/2.3.0).

## Installation 

### Ensure proper aibench version
```
git clone https://github.com/RedisAI/aibench
git checkout v0.2.0
```

### Docker Installation -- Download the TensorFlow Serving ARM Docker image and repo

```bash
docker pull emacski/tensorflow-serving:2.3.0
cd $GOPATH/src/github.com/RedisAI/aibench

# Location of credit card fraud model
TESTDATA="$(pwd)/tests/models/tensorflow/reference"

# Start TensorFlow Serving container and open the GRPC and REST API ports
docker run -t --rm -p 8500:8500 -p 8501:8501 \
    -v "$TESTDATA:/models/financialNet" \
    -e MODEL_NAME=financialNet -e TF_CPP_MIN_VLOG_LEVEL=1 \
    -d emacski/tensorflow-serving:2.3.0

# Query the model using the predict API
curl --data @$(pwd)/tests/models/tensorflow/tensorflow_serving_inference_payload.json -X POST http://localhost:8501/v1/models/financialNet:predict
```
Expected output:
```
~/aibench# curl --data @$(pwd)/tests/models/tensorflow/tensorflow_serving_inference_payload.json -X POST http://localhost:8501/v1/models/financialNet:predict
{
    "outputs": [
        [
            0.87545836,
            0.12454161
        ]
    ]
}
```


## Benchmarking inference performance -- TensorFlow Serving on ARM and Redis Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/RedisAI/aibench#reference-data-loading). 

Once the data is loaded,
just use the corresponding `aibench_run_inference_tensorflow_serving` binary for the DL Solution
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
./scripts/run_inference_tensorflow_serving.sh
```
