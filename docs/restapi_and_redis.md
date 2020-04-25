# aibench Supplemental Guide: DL REST API and Redis


### Benchmarking inference performance -- DL REST API and Redis Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/RedisAI/aibench#reference-data-loading). 

Once the data is loaded,
just use the corresponding `aibench_run_inference_flask_tensorflow` binary for the DL Solution
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/RedisAI/aibench
./scripts/run_inference_flask_tensorflow.sh
```


#### Sequence diagram - DL REST API Solution

The following diagram illustrates the sequence of requests made for each inference.

![Sequence diagram - DL REST API Solution][aibench_client_restapi]

[aibench_client_restapi]: ./aibench_client_restapi.png

---

### Local Installation -- with wsgi + flask for Development mode

```bash
cd $GOPATH/src/github.com/RedisAI/aibench/tests/servers/flask

# Install requirements
pip install -r requirements

# set environment variable with location of credit card fraud model
export TF_MODEL_PATH=$GOPATH/src/github.com/RedisAI/aibench/tests/models/tensorflow/creditcardfraud.pb
# Start WSGI+Flask+TF Backend REST API serving
python3 server.py
```

### Production Installation -- Install Guinicorn, Flask, and Tensorflow backend on production VM


```bash
cd $GOPATH/src/github.com/RedisAI/aibench/tests/servers/flask

# Install requirements
apt install gunicorn -y
pip install -r requirements.txt

export TF_MODEL_PATH=$GOPATH/src/github.com/RedisAI/aibench/tests/models/tensorflow/creditcardfraud.pb
gunicorn --workers=48 --threads 48 -b 0.0.0.0:8000  --log-level error --daemon server:app
```
