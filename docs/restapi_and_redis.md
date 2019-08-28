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
         -max-queries 100000 -workers 16 -print-interval 25000 \
         -restapi-host localhost:8000 \
         -restapi-request-uri /v2/predict \
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
curl  -H "Content-Type: application/json" \
      --data @$TESTDATA/models/tensorflow/tensorflow_serving_inference_payload.json \
      -X POST http://localhost:8000/predict

# Returns => {
               "outputs": [
                 [
                   0.9055531620979309, 
                   0.09444686770439148
                 ]
               ]
             }

```

### Local Installation -- with tensorwerk flask-optim-cpu Docker image

```bash
docker pull tensorwerk/raibenchmarks:flask-optim-cpu
cd $GOPATH/src/github.com/filipecosta90/aibench

# Location of credit card fraud model
TESTDATA="$(pwd)/tests"

# Start Guinicorn+Flask+ TF Backend container and open the REST API port
docker run --read-only -v $TESTDATA/models/tensorflow:/root/data \
    --read-only -v $TESTDATA/servers/flask:/root \
    -p 8000:8000 --name server -d --rm tensorwerk/raibenchmarks:flask-optim-cpu

# Query the model using the predict API
curl  -H "Content-Type: application/json" \
      --data @$TESTDATA/models/tensorflow/tensorflow_serving_inference_payload.json \
      -X POST http://localhost:8000/predict

# Returns => {
               "outputs": [
                 [
                   0.9055531620979309, 
                   0.09444686770439148
                 ]
               ]
             }

```


### Production Installation -- Install Guinicorn, Flask, and Tensorflow backend on production VM

TBD
