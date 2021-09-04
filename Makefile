APP_NAME=kfserving-inference-batch-client

build:
	GOOS=linux GOARCH=amd64 go build -o kfserving-inference-client
	docker build -t $(APP_NAME) .

build-nc:
	docker build --no-cache -t $(APP_NAME) .

go run . -i example/make/assert/input-data.txt -o example/make/assert/output-data.txt -m sklearn -host localhost:5001 -w 1
