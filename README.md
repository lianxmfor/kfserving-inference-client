# KFServing Inference Client

A Go re-implementation of [seldon-batch-processor](https://github.com/SeldonIO/seldon-core/blob/master/python/seldon_core/batch_processor.py). See [docs](https://docs.seldon.io/projects/seldon-core/en/stable/servers/batch.html) to understand its usage.

The main reason why we choose to re-implement is because the original implementation follow Seldon protocol, while we need KFServing V2 Protocol in order to use [MLServer](https://github.com/SeldonIO/MLServer) as the inference backed.

## man 

```sh
$ ./kfserving-inference-client -h
Usage of ./kfserving-inference-client:
  -host string
    	The hostname for the seldon model to send the request to, which can be the ingress of the Seldon model or the service itself
  -i string
    	The local filestore path where the input file with the data to process is located
  -m string
    	model name
  -o string
    	The local filestore path where the output file should be written with the outputs of the batch processing
  -u int
    	Batch size greater than 1 can be used to group multiple predictions into a single request. (default 100)
  -w int
    	The number of parallel request processor workers to run for parallel processing (default 100)
```

## Build Docker Image

$ make build
