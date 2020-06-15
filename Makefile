# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# DOCKER
DOCKER_APP_NAME=aibench
DOCKER_ORG=redisbench
DOCKER_REPO:=${DOCKER_ORG}/${DOCKER_APP_NAME}
DOCKER_IMG:="$(DOCKER_REPO):$(DOCKER_TAG)"
DOCKER_LATEST:="${DOCKER_REPO}:latest"

.PHONY: all generators loaders runners
all: generators loaders runners

generators: aibench_generate_data aibench_generate_data_vision

loaders: aibench_load_data

runners: aibench_run_inference_redisai aibench_run_inference_redisai_vision aibench_run_inference_torchserve aibench_run_inference_flask_tensorflow aibench_run_inference_tensorflow_serving

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

data: generators
	./scripts/generate_data.sh

data-vision: generators
	./scripts/generate_data_vision.sh

aibench_%: $(wildcard ./cmd/$@/*.go)
	$(GOGET) ./cmd/$@
	$(GOBUILD) -o bin/$@ ./cmd/$@
	$(GOINSTALL) ./cmd/$@

# DOCKER TASKS
# Build the container
docker-build:
	docker build -t $(DOCKER_APP_NAME):latest -f  docker/Dockerfile .

# Build the container without caching
docker-build-nc:
	docker build --no-cache -t $(DOCKER_APP_NAME):latest -f docker/Dockerfile .

# Make a release by building and publishing the `{version}` ans `latest` tagged containers to ECR
docker-release: docker-build-nc docker-publish

# Docker publish
docker-publish: docker-publish-latest

## login to DockerHub with credentials found in env
docker-repo-login:
	docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}

## Publish the `latest` tagged container to ECR
docker-publish-latest: docker-tag-latest
	@echo 'publish latest to $(DOCKER_REPO)'
	docker push $(DOCKER_LATEST)

# Docker tagging
docker-tag: docker-tag-latest

## Generate container `{version}` tag
docker-tag-latest:
	@echo 'create tag latest'
	docker tag $(DOCKER_APP_NAME) $(DOCKER_LATEST)
