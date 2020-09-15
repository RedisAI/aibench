# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get -v
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

.PHONY: all generators loaders runners

all: generators loaders runners

financial: aibench_generate_data aibench_load_data aibench_run_inference_redisai aibench_run_inference_torchserve aibench_run_inference_flask_tensorflow aibench_run_inference_tensorflow_serving

generators: aibench_generate_data aibench_generate_data_vision

loaders: aibench_load_data

runners: aibench_run_inference_redisai aibench_run_inference_redisai_vision aibench_run_inference_triton_vision aibench_run_inference_torchserve aibench_run_inference_flask_tensorflow aibench_run_inference_tensorflow_serving

fmt:
	go fmt ./...

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
	$(GOGET) ./cmd/$@
	$(GOBUILD) ./cmd/$@
	$(GOINSTALL) ./cmd/$@

#####################
###### helpers ######
#####################

data: generators
	./scripts/generate_data.sh

load-financial: loaders
	./scripts/load_tensors_redis.sh

data-vision: generators
	./scripts/generate_data_vision.sh


