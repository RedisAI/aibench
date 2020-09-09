# aibench Supplemental Guide: DL REST API and Redis


### Benchmarking inference performance -- DL REST API and Redis Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/RedisAI/aibench#reference-data-loading), 
and ensure that the model server is installed and running as explained bellow. If you already have your environment setted up jump to Running DL solution part.



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
