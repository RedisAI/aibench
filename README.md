[![license](https://img.shields.io/github/license/RedisAI/aibench.svg)](https://github.com/RedisAI/aibench)
[![CircleCI](https://circleci.com/gh/RedisAI/aibench.svg?style=svg)](https://circleci.com/gh/RedisAI/aibench)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/RedisAI/redisai-go)
[![Forum](https://img.shields.io/badge/Forum-RedisAI-blue)](https://forum.redislabs.com/c/modules/redisai)
[![Discord](https://img.shields.io/discord/697882427875393627?style=flat-square)](https://discord.gg/rTQm7UZ)

# aibench
This repo contains code for benchmarking deep learning solutions,
including RedisAI.
This code is based on a fork of work initially made public by TSBS
at https://github.com/timescale/tsbs.

###  Current DL solutions supported:

- [RedisAI](https://redisai.io): an AI serving engine for real-time applications built by Redis Labs and Tensorwerk, seamlessly plugged into ​Redis.
- [Nvidia Triton Inference Server](https://docs.nvidia.com/deeplearning/triton-inference-server): An open source inference serving software that lets teams deploy trained AI models from any framework (TensorFlow, TensorRT, PyTorch, ONNX Runtime, or a custom framework), from local storage or Google Cloud Platform or AWS S3 on any GPU- or CPU-based infrastructure.
- [TorchServe](https://pytorch.org/serve/): built and maintained by Amazon Web Services (AWS) in collaboration with Facebook, TorchServe is available as part of the PyTorch open-source project.
- [Tensorflow Serving](https://www.tensorflow.org/tfx/guide/serving): a high-performance serving system, wrapping TensorFlow and maintained by Google.
- [Common REST API serving](https://redisai.io): a common DL production grade setup with Gunicorn (a Python WSGI HTTP server) communicating with Flask through a WSGI protocol, and using TensorFlow as the backend.

### Current use cases

Currently, aibench supports two use cases: 
 - **creditcard-fraud [[details here](docs/creditcard-fraud-benchmark/description.md)]**: from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) with the extension of reference data. This use-case aims to detect a fraudulent transaction based on anonymized credit card transactions and reference data. 
 
 - **vision-image-classification[[details here](docs/vision-image-classification-benchmark/description.md)]**: an image-focused use-case that uses one network “backbone”: MobileNet V1, which can be considered as one of the standards by the AI community. To assess inference performance we’re recurring to COCO 2017 validation dataset (a large-scale object detection, segmentation, and captioning dataset).
### Current DL solutions supported per use case:

| Use case/Inference Server      | model | RedisAI  | TensorFlow Serving | Torch Serve | Nvidia Triton | Rest API |
|--------------------------------|----------|----------|--------------------|-------------|---------------|----------|
| Vision Benchmark (CPU/GPU) ([details](docs/vision-image-classification-benchmark/description.md)) | [mobilenet-v1 (224_224)](https://zenodo.org/record/2269307/files/mobilenet_v1_1.0_224.tgz)| :heavy_check_mark: | Not supported          | Not supported    | :heavy_check_mark:     | Not supported |
| Fraud Benchmark (CPU) ([details](docs/creditcard-fraud-benchmark/description.md)) |   [Non standard Kaggle Model](https://www.kaggle.com/mlg-ulb/creditcardfraud) with the extension of reference data    | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/redisai.md) | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/tf_serving_and_redis.md)           | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/torchserve_and_redis.md)    | Not supported    | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/restapi_and_redis.md) |



### Installation

The easiest way to get and install the go benchmark programs is to use `go get` and then issuing make:
```
# Fetch aibench and its dependencies
go get github.com/RedisAI/aibench
cd $GOPATH/src/github.com/RedisAI/aibench

make
```

### Blogs/White-papers that reference this tool

- [1] [Announcing RedisAI 1.0: AI Serving Engine for Real-Time Applications](https://redislabs.com/blog/redisai-ai-serving-engine-for-real-time-applications/), May 19, 2020
