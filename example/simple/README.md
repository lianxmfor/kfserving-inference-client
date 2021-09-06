# Deploy LightGBM models using seldon batch inference and argo workflow

Since seldon-batch-processor does not support the kfserving inference protocol and the Inference server I built for lightGBM only supports the kfserving inference protocol, so I developed the kfserving-inference-client to replace seldon-batch-processor.

The lightgbm example model [simple](https://github.com/microsoft/LightGBM/blob/master/examples/python-guide/simple_example.py) is used here for demonstration purposes.


## Setup

1. [Install Seldon core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html)
2. [Install Argo Workflows](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html)
3. [Install MinIO](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html)

## Register lightGBM Inference Server

seldon core does not currently have an inference server that supports the lightGBM model,so I registered the lightgbm-enabled inference server I build with seldon. can register our Inference server with seldon core by referring to [document](https://docs.seldon.io/projects/seldon-core/en/latest/servers/custom.html#adding-a-new-inference-server)

``` json
 {
   ...
  "LightGBM_SERVER": {
      "protocols": {
        "kfserving":{
          "defaultImageVersion": "v0.0.3",
	        "image": "1034552569/seldon-lightgbm"
	     }
      }
  }
}
```


## Upload the file to our minio

``` shell
$ mc mb minio-seldon/data/simple
$ mc cp assert/input-data.csv minio-seldon/data/simple                                

```

## Create oss config

``` shell
$ kubectl -n argo create -f rclone-config-secret.yaml
```

## Create Argo Workflow

``` shell
$ argo -n argo submit workflow.yaml
$ argo -n argo get  seldon-batch-process-simple 
Name:                seldon-batch-process-simple
Namespace:           argo
ServiceAccount:      default
Status:              Succeeded
Conditions:          
 PodRunning          False
 Completed           True
Created:             Mon Sep 06 17:39:24 +0800 (1 hour ago)
Started:             Mon Sep 06 17:39:24 +0800 (1 hour ago)
Finished:            Mon Sep 06 17:42:01 +0800 (1 hour ago)
Duration:            2 minutes 37 seconds
Progress:            6/6
ResourcesDuration:   57s*(1 cpu),57s*(100Mi memory)

STEP                            TEMPLATE                         PODNAME                                 DURATION  MESSAGE
 ✔ seldon-batch-process-simple  seldon-batch-process                                                                 
 ├───✔ create-seldon-resource   create-seldon-resource-template  seldon-batch-process-simple-3967243303  7s          
 ├───✔ wait-seldon-resource     wait-seldon-resource-template    seldon-batch-process-simple-143792717   42s         
 ├───✔ download-object-store    download-object-store-template   seldon-batch-process-simple-1244423636  12s         
 ├───✔ process-batch-inputs     process-batch-inputs-template    seldon-batch-process-simple-2768547645  17s         
 ├───✔ upload-object-store      upload-object-store-template     seldon-batch-process-simple-2531549797  14s         
 └───✔ delete-seldon-resource   delete-seldon-resource-template  seldon-batch-process-simple-204379243   19s         
```


## Check output file

``` shell
$ mc ls minio-seldon/data/simple/                                                    
[2021-09-06 19:23:11 CST]   581B input-data.csv
[2021-09-06 18:50:51 CST]   221B output-data-968000bb-4810-4673-b201-f35dd889e38e.csv
$ mc cat minio-seldon/data/simple/output-data-968000bb-4810-4673-b201-f35dd889e38e.csv
1,0.46417540311813354
2,0.46417540311813354
3,0.46417540311813354
4,0.46417540311813354
5,0.46417540311813354
6,0.46417540311813354
7,0.46417540311813354
8,0.46417540311813354
9,0.46417540311813354
10,0.46417540311813354
```
