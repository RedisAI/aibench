# aibench Supplemental Guide: DL REST API and Redis


### Benchmarking inference performance -- DL REST API and Redis Benchmark Go program

To measure inference performance in aibench, you first need to load
the data using the instructions in the overall [Reference data Loading section](https://github.com/filipecosta90/aibench#reference-data-loading). 

Once the data is loaded,
just use the corresponding `aibench_run_inference_flask_tensorflow` binary for the DL Solution
being tested:

```bash
# make sure you're on the root project folder
cd $GOPATH/src/github.com/filipecosta90/aibench
cat /tmp/aibench_generate_data-creditcard-fraud.dat.gz \
        | gunzip \
        | aibench_run_inference_flask_tensorflow \
         -workers 8 \
         -burn-in 10 -max-queries 100010 \
         -print-interval 10000 -reporting-period 1000ms \
         -restapi-host localhost:8000 \
         -restapi-request-uri /v2/predict \
         -restapi-read-timeout 30s \
         -redis-host localhost:6379
```


#### Sequence diagram - DL REST API Solution

The following diagram illustrates the sequence of requests made for each inference.

![Sequence diagram - DL REST API Solution][aibench_client_restapi]

[aibench_client_restapi]: ./aibench_client_restapi.png

---

### Local Installation -- with wsgi + flask for Development mode

```bash
cd $GOPATH/src/github.com/filipecosta90/aibench/tests/servers/flask

# Install requirements
pip install -r requirements

# set environment variable with location of credit card fraud model
export TF_MODEL_PATH=$GOPATH/src/github.com/filipecosta90/aibench/tests/models/tensorflow/creditcardfraud.pb

# Start WSGI+Flask+TF Backend REST API serving
python3 server.py

# Query the model using the predict API
# TBD
```

### Production Installation -- Install Guinicorn, Flask, and Tensorflow backend on production VM

TBD
