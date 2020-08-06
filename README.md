[![license](https://img.shields.io/github/license/RedisAI/aibench.svg)](https://github.com/RedisAI/aibench)
[![Forum](https://img.shields.io/badge/Forum-RedisAI-blue)](https://forum.redislabs.com/c/modules/redisai)
[![Gitter](https://badges.gitter.im/RedisLabs/RedisAI.svg)](https://gitter.im/RedisLabs/RedisAI?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

# aibench
This repo contains code for benchmarking deep learning solutions,
including RedisAI.
This code is based on a fork of work initially made public by TSBS
at https://github.com/timescale/tsbs.

## Current use cases

Currently, aibench supports two use cases: 
 - creditcard-fraud from [Kaggle](https://www.kaggle.com/mlg-ulb/creditcardfraud) with the extension of reference data. This use-case aims to detect a fraudulent transaction based on anonymized credit card transactions and reference data. 
 
 
 - vision-image-classification, an image-focused use-case that uses one network “backbone”: MobileNet V1, which can be considered as one of the standards by the AI community. To assess inference performance we’re recurring to COCO 2017 validation dataset (a large-scale object detection, segmentation, and captioning dataset).
## Current DL solutions supported:

| Use case/Inference Server      | model | RedisAI  | TensorFlow Serving | Torch Serve | Nvidia Triton | Rest API |
|--------------------------------|----------|----------|--------------------|-------------|---------------|----------|
| Vision Benchmark (CPU/GPU) ([details](docs/vision-image-classification-benchmark/description.md)) | mobilenet-v1 (224_224 )| :heavy_check_mark: | Not supported          | Not supported    | :heavy_check_mark:     | Not supported |
| Fraud Benchmark (CPU) ([details](docs/creditcard-fraud-benchmark/description.md)) |   [Non standard Kaggle Model](https://www.kaggle.com/mlg-ulb/creditcardfraud) with the extension of reference data    | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/redisai.md) | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/tf_serving_and_redis.md)           | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/torchserve_and_redis.md)    | Not supported    | :heavy_check_mark: [docs](docs/creditcard-fraud-benchmark/restapi_and_redis.md) |

