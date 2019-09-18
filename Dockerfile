FROM golang:1.13 AS builder

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/RedisAI/aibench
COPY . ./
RUN cd $GOPATH/src/github.com/RedisAI/aibench/cmd && CGO_ENABLED=0 GOOS=linux go get ./...
RUN cd $GOPATH/src/github.com/RedisAI/aibench/cmd/aibench_generate_data && CGO_ENABLED=0 GOOS=linux go build -o /tmp/aibench_generate_data
RUN cd $GOPATH/src/github.com/RedisAI/aibench/cmd/aibench_load_data  && CGO_ENABLED=0 GOOS=linux go build -o /tmp/aibench_load_data
RUN cd $GOPATH/src/github.com/RedisAI/aibench/cmd/aibench_run_inference_redisai  && CGO_ENABLED=0 GOOS=linux go build -o /tmp/aibench_run_inference_redisai
RUN cd $GOPATH/src/github.com/RedisAI/aibench/cmd/aibench_run_inference_tensorflow_serving  && CGO_ENABLED=0 GOOS=linux go build -o /tmp/aibench_run_inference_tensorflow_serving
RUN cd $GOPATH/src/github.com/RedisAI/aibench/cmd/aibench_run_inference_flask_tensorflow  && CGO_ENABLED=0 GOOS=linux go build -o /tmp/aibench_run_inference_flask_tensorflow

FROM golang:1.13.0-alpine3.10
COPY --from=builder /tmp/aibench_generate_data ./
COPY --from=builder /tmp/aibench_load_data ./
COPY --from=builder /tmp/aibench_run_inference_redisai ./
COPY --from=builder /tmp/aibench_run_inference_tensorflow_serving ./
COPY --from=builder /tmp/aibench_run_inference_flask_tensorflow ./
COPY docker_entrypoint.sh ./
RUN chmod 751 docker_entrypoint.sh
ENTRYPOINT ["./docker_entrypoint.sh"]