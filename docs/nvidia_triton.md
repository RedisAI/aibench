# aibench Supplemental Guide: Nvidia Triton Inference Server


## Model Repository
More info at https://github.com/NVIDIA/triton-inference-server/blob/master/docs/model_repository.rst
### TensorFlow Models

TensorFlow saves trained models in one of two ways: GraphDef or SavedModel. Triton supports both formats. Once you have a trained model in TensorFlow, you can save it as a GraphDef directly or convert it to a GraphDef by using a script like freeze_graph.py, or save it as a SavedModel using a SavedModelBuilder or tf.saved_model.simple_save. If you use the Estimator API you can also use Estimator.export_savedmodel.

A TensorFlow GraphDef is a single file that by default must be named model.graphdef. A minimal model repository for a single TensorFlow GraphDef model would look like:
```
<model-repository-path>/
  <model-name>/
    config.pbtxt
    1/
      model.graphdef
```

Here is the config.pbtxt file
```
name: "mobilenet_v1_100_224_NxHxWxC"
platform: "tensorflow_graphdef"
max_batch_size: 1
input [
   {
      name: "inputs"
      data_type: TYPE_FP32
      format: FORMAT_NCHW
      dims: [ 3, 224, 224 ]
   }
]
output [
   {
      name: "MobilenetV1/Predictions/Reshape_1"
      data_type: TYPE_FP32
      dims: [ 1001 ]
   }
]
```

## Installation 

### Using A Prebuilt Docker Container

Use docker pull to get the Triton Inference Server container from NGC:
```
docker pull nvcr.io/nvidia/tritonserver:20.03-py3
```

CPU only 
```
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
docker run --rm -p8000:8000 -p8001:8001 -v$(pwd)/tests/models/triton-tensorflow-model-repository:/models nvcr.io/nvidia/tritonserver:20.03.1-py3 trtserver --model-store=/models
```

GPU capable

```
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
nvidia-docker run --rm -p8000:8000 -p8001:8001 -v$(pwd)/tests/models/triton-tensorflow-model-repository:/models nvcr.io/nvidia/tritonserver:20.03.1-py3 trtserver --model-store=/models
```


#### Query the model using the predict API

TBD

### Production Installation 

TBD
