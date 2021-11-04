APP_NAME := 1034552569/kfserving-inference-batch-client
APP_TAG  := v12
IMAGE    := $(APP_NAME):$(APP_TAG)

build:
	@go build -o kfserving-inference-client

docker:
	@go build -o kfserving-inference-client
	@docker build -t $(IMAGE) .
