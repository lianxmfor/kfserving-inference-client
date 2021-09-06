APP_NAME :=1034552569/kfserving-inference-batch-client
APP_TAG  :=v3
IMAGE    := $(APP_NAME):$(APP_TAG)

build:
	GOOS=linux GOARCH=amd64 go build -o kfserving-inference-client
	docker build -t $(IMAGE) .

build-nc:
	docker build --no-cache -t $(APP_NAME) .
