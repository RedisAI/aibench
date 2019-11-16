# aibench Supplemental Guide: Tensorflow Serving and Redis

### Benchmarking inference performance -- TFServing and Redis Benchmark Go program

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

#### Sequence diagram - Tensorflow Serving and Redis Solution

The following diagram illustrates the sequence of requests made for each inference.

![Sequence diagram - Tensorflow Serving and Redis Solution][aibench_client_tfserving]

[aibench_client_tfserving]: ./aibench_client_tfserving.png

---

## Installation

### Prerequisites

#### Go support for Protocol buffers (Google's data interchange format)
                                                                                                        
                                                             
 
 Tensorflow Serving is written in C++, exposing a gRPC server that talks Protobuf.
  
  We consider you have protobuf installed and are only installing Go language support for it. If not please follow the official documentation [here](
                                                                                    https://github.com/protocolbuffers/protobuf/).
                                                                                    
                                                                                    
In order to make a Go client, we must compile the protobuf files first to generate all the boilerplate code for Go. For that we need protoc-gen-go package.
The simplest way to get it, is to run

 ```bash
 go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
 go get -u google.golang.org/grpc
 ```

 
 The compiler plugin, protoc-gen-go, will be installed in $GOPATH/bin unless $GOBIN is set. It must be in your $PATH for the protocol compiler, protoc, to find it.


```bash
cd $GOPATH/src/github.com/RedisAI/aibench
mkdir -p tmp && cd tmp
git clone -b r1.7 --depth 1 https://github.com/tensorflow/serving.git
git clone -b r1.7 --depth 1 https://github.com/tensorflow/tensorflow.git

mkdir -p vendor
PROTOC_OPTS=' -I serving -I tensorflow --go_out=plugins=grpc:vendor'
protoc $PROTOC_OPTS serving/tensorflow_serving/apis/*.proto
protoc $PROTOC_OPTS serving/tensorflow_serving/config/*.proto
protoc $PROTOC_OPTS serving/tensorflow_serving/util/*.proto
protoc $PROTOC_OPTS serving/tensorflow_serving/sources/storage_path/*.proto
protoc $PROTOC_OPTS tensorflow/tensorflow/core/framework/*.proto
protoc $PROTOC_OPTS tensorflow/tensorflow/core/example/*.proto
protoc $PROTOC_OPTS tensorflow/tensorflow/core/lib/core/*.proto
protoc $PROTOC_OPTS tensorflow/tensorflow/core/protobuf/{saver,meta_graph}.proto

# move vendor folder to $GOPATH/src/github.com/RedisAI/aibench
rm -rf $GOPATH/src/github.com/RedisAI/aibench/vendor
mv vendor $GOPATH/src/github.com/RedisAI/aibench/.

# remove tmp dir
cd .. && rm -rf tmp
 ```
 
## Installation 

### Local Installation -- Download the TensorFlow Serving Docker image and repo

```bash
docker pull tensorflow/serving
cd $GOPATH/src/github.com/RedisAI/aibench

# Location of credit card fraud model
TESTDATA="$(pwd)/tests/models/tensorflow"

# Start TensorFlow Serving container and open the GRPC and REST API ports
docker run -t --rm -p 8500:8500 -p 8501:8501 \
    -v "$TESTDATA:/models/financialNet" \
    -e MODEL_NAME=financialNet -e TF_CPP_MIN_VLOG_LEVEL=1 \
    -d tensorflow/serving 

# Query the model using the predict API
curl --data @$TESTDATA/tensorflow_serving_inference_payload.json -X POST http://localhost:8501/v1/models/financialNet:predict

# Returns => { "outputs": [[ 0.889327943, 0.110672079 ]] }
```

### Production Installation -- Install TensorFlow Serving on production VM

```bash
tensorflow_model_server --inter_op_parallelism_threads=4 --model_name=financialNet --model_base_path=$GOPATH/src/github.com/RedisAI/aibench/tests/models/tensorflow
```
