# DLBench Supplemental Guide: TFServing and Redis

TBD



### Prerequisites

#### Tensorflow Serving Installation

TBD


#### Go support for Protocol buffers (Google's data interchange format)
                                                                                                        
                                                             
 
 Tensorflow Serving is written in C++, exposing a gRPC server that talks Protobuf.
  
  We consider you have protobuf installed and are only installing Go language support for it. If not please follow the official documentation [here](
                                                                                    https://github.com/protocolbuffers/protobuf/).
                                                                                    
                                                                                    
In order to make a Go client, we must compile the protobuf files first to generate all the boilerplate code for Go. For that we need protoc-gen-go package.
The simplest way to get it, is to run

 ```bash
 $ go get -u github.com/golang/protobuf/protoc-gen-go
 ```

 
 The compiler plugin, protoc-gen-go, will be installed in $GOPATH/bin unless $GOBIN is set. It must be in your $PATH for the protocol compiler, protoc, to find it.


```bash
$ mkdir tmp 
$ cd tmp
$ git clone --branch 1.14.0 --depth 1 https://github.com/tensorflow/serving.git
$ git clone --branch v1.13.2 --depth 1 https://github.com/tensorflow/tensorflow.git

$ mkdir -p vendor
$ PROTOC_OPTS=' -I serving -I tensorflow --go_out=plugins=grpc:vendor'
$ protoc $PROTOC_OPTS serving/tensorflow_serving/apis/*.proto
$ protoc $PROTOC_OPTS serving/tensorflow_serving/config/*.proto
$ protoc $PROTOC_OPTS serving/tensorflow_serving/util/*.proto
$ protoc $PROTOC_OPTS serving/tensorflow_serving/sources/storage_path/*.proto
$ protoc $PROTOC_OPTS tensorflow/tensorflow/core/framework/*.proto
$ protoc $PROTOC_OPTS tensorflow/tensorflow/core/example/*.proto
$ protoc $PROTOC_OPTS tensorflow/tensorflow/core/lib/core/*.proto
$ protoc $PROTOC_OPTS tensorflow/tensorflow/core/protobuf/{saver,meta_graph}.proto

# move vendor folder to $GOPATH/src/github.com/filipecosta90/dlbench
$ mv vendor $GOPATH/src/github.com/filipecosta90/dlbench/.
# remove tmp dir
$ rm -rf tmp
 ```
 
 
