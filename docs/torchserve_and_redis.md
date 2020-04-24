# aibench Supplemental Guide: TorchServe and Redis

### Benchmarking inference performance -- TFServing and Redis Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/RedisAI/aibench#reference-data-loading). 

Once the data is loaded,
just use the corresponding `aibench_run_inference_torchserve` binary for the DL Solution
being tested, or an helper script that will enable to quickly benchmark your model server limits:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
./scripts/run_inference_torchserve.sh
```

### Production Deployment Steps 

#### Once in a time setup
```bash
# Ensure openjdk installed 
apt-get install openjdk-11-jdk -y

# Install torch 
python3 -m pip install torch torchtext torchvision sentencepiece nvidia-ml-py3

# Install TorchServe and the model archiver
python3 -m  pip install torchserve torch-model-archiver
```

#### Store the financial model
```bash
# make sure you're on the torchserve folder
cd $GOPATH/src/github.com/RedisAI/aibench/tests/models/torch
torch-model-archiver --model-name financialNetTorch --version 1 --serialized-file torchFraudNetWithRef.pt --handler handler_financialNet.py
```

#### Start TorchServe to serve the model
```bash
# make sure you're on the torchserve folder
cd $GOPATH/src/github.com/RedisAI/aibench/tests/models/torch
torchserve --start --model-store . --models financial=financialNetTorch.mar --ts-config config.properties --log-config log4j.properties
```

#### config.properties file
```bash
async_logging=true
inference_address=http://0.0.0.0:8080
management_address=http://0.0.0.0:8081
```

#### log4j.properties file
```bash
log4j.logger.com.amazonaws.ml.ts = WARN
```