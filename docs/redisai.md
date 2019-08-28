# aibench Supplemental Guide: RedisAI


### Benchmarking inference performance -- RedisAI Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/filipecosta90/aibench#reference-data-loading). 

Once the data is loaded,
just use the corresponding `aibench_run_inference_redisai` binary for the DL Solution
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/aibench
cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz \
        | gunzip \
        | aibench_run_inference_redisai \
         -max-queries 200000 -workers 16 -print-interval 100000 \
         -model financialNet \
         -host redis://127.0.0.1:6379 
```

#### Sequence diagram - RedisAI Solution

The following diagram illustrates the sequence of requests made for each inference.


![Sequence diagram - RedisAI Solution][aibench_client_redisai]

[aibench_client_redisai]: ./aibench_client_redisai.png





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
cd $GOPATH/src/github.com/filipecosta90/aibench
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

# move vendor folder to $GOPATH/src/github.com/filipecosta90/aibench
rm -rf $GOPATH/src/github.com/filipecosta90/aibench/vendor
mv vendor $GOPATH/src/github.com/filipecosta90/aibench/.

# remove tmp dir
cd .. && rm -rf tmp
 ```
 
## Installation 

### Local Installation -- Download the RedisAI Docker image

```bash
docker pull redisai/redisai

# Start RedisAI container 
docker run -t --rm -p 6379:6379 
```

### Production Installation -- Install RedisAI on production VM

TBD
