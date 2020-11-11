# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get -v
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build time variables
ifeq ($(GIT_SHA),)
GIT_SHA:=$(shell git rev-parse HEAD)
endif

ifeq ($(GIT_DIRTY),)
GIT_DIRTY:=$(shell git diff --no-ext-diff 2> /dev/null | wc -l)
endif

.PHONY: all generators loaders runners

all: generators loaders runners

redisai: aibench_generate_data aibench_generate_data_vision aibench_load_data aibench_run_inference_redisai aibench_run_inference_redisai_vision

financial: aibench_generate_data aibench_load_data aibench_run_inference_redisai aibench_run_inference_torchserve aibench_run_inference_flask_tensorflow aibench_run_inference_tensorflow_serving

generators: aibench_generate_data aibench_generate_data_vision

loaders: aibench_load_data

runners: aibench_run_inference_redisai aibench_run_inference_redisai_vision aibench_run_inference_triton_vision aibench_run_inference_torchserve aibench_run_inference_flask_tensorflow aibench_run_inference_tensorflow_serving

fmt:
	$(GOFMT) ./...

checkfmt:
	@echo 'Checking gofmt';\
 	bash -c "diff -u <(echo -n) <(gofmt -d .)";\
	EXIT_CODE=$$?;\
	if [ "$$EXIT_CODE"  -ne 0 ]; then \
		echo '$@: Go files must be formatted with gofmt'; \
	fi && \
	exit $$EXIT_CODE

get:
	$(GOGET) -t -v ./...

test: get
	$(GOFMT) ./...
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

aibench_%: $(wildcard ./cmd/$@/*.go) ./inference/*.go
	#$(GOGET) ./cmd/$@
	$(GOBUILD) -ldflags="-X 'main.GitSHA1=$(GIT_SHA)' -X 'main.GitDirty=$(GIT_DIRTY)'" ./cmd/$@
	$(GOINSTALL) ./cmd/$@

#####################
###### helpers ######
#####################

load-fraud: loaders
	./scripts/load_tensors_redis.sh

data-fraud-ci: generators
	DEBUG=1 NUM_INFERENCES=100000 ./scripts/generate_data.sh

data-vision-ci: generators
	DEBUG=1 ./scripts/generate_data_vision.sh

bench-fraud-ci:
	SLEEP_BETWEEN_RUNS=0 CLIENTS_STEP=16 MIN_CLIENTS=0 MAX_CLIENTS=16 NUM_INFERENCES=100000 RUNS_PER_VARIATION=1 ./scripts/run_inference_redisai_fraud.sh

bench-vision-ci:
	./scripts/load_models_mobilenet_redisai.sh
	SLEEP_BETWEEN_RUNS=0 CLIENTS_STEP=16 MIN_CLIENTS=0 MAX_CLIENTS=16 NUM_INFERENCES=5000 RUNS_PER_VARIATION=1 ./scripts/run_inference_redisai_vision.sh

