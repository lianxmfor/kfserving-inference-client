package main

import (
	"context"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"kfserving-inference-client/inference"
	"kfserving-inference-client/mapping"

	"github.com/spf13/cast"
	"google.golang.org/grpc"
)

var (
	inputDataPath  string
	outputDataPath string
	host           string
	modelName      string
	mappingPath    string
	worker         int
	batchSize      int64
)

func init() {
	InitKFServingGrpcClient(time.Second * 10)

	flag.StringVar(&inputDataPath, "i", "", "The local filestore path where the input file with the data to process is located")
	flag.StringVar(&outputDataPath, "o", "", "The local filestore path where the output file should be written with the outputs of the batch processing")
	flag.StringVar(&host, "host", "", "The hostname for the seldon model to send the request to, which can be the ingress of the Seldon model or the service itself ")
	flag.StringVar(&modelName, "m", "", "model name")
	flag.StringVar(&mappingPath, "mapping_path", ".", "The feature mapping csv file path")
	flag.IntVar(&worker, "w", 100, "The number of parallel request processor workers to run for parallel processing")
	flag.Int64Var(&batchSize, "u", 100, "Batch size greater than 1 can be used to group multiple predictions into a single request.")
}

func main() {
	flag.Parse()

	mapping.Init(mappingPath)

	in := make(chan request, worker)
	out := make(chan response, worker)

	ctx, cancel := context.WithCancel(context.Background())

	go getRequestFromFile(cancel, inputDataPath, in)

	go startRequest(ctx, worker, in, out)

	writeResponseToFile(outputDataPath, out)
}

func startRequest(ctx context.Context, worker int, in <-chan request, out chan<- response) {
	var wait sync.WaitGroup
	for i := 0; i < worker; i++ {
		wait.Add(1)
		go requestWorker(ctx, &wait, in, out)
	}
	wait.Wait()
	close(out)
}

func requestWorker(ctx context.Context, wait *sync.WaitGroup, in <-chan request, out chan<- response) {
	defer wait.Done()

	doRequest := func(r *RequestChunk, conn *grpc.ClientConn) {
		res, err := kfServingGrpcClient.Inference(context.Background(), host, &inference.ModelInferRequest{
			ModelName: modelName,
			Inputs: []*inference.ModelInferRequest_InferInputTensor{
				{
					Shape:    r.Shape(),
					Datatype: "FP64",
					Contents: &inference.InferTensorContents{
						Fp64Contents: r.Tensor,
					},
				},
			},
		}, conn)
		if err != nil {
			panic(err)
		}

		for i, content := range res.Outputs[0].Contents.Fp64Contents {
			out <- response{
				EntityKey:         r.EntityKey[i],
				InferenceResponse: cast.ToFloat64(content),
			}
		}
	}

	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(time.Second*5), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var (
		chunk = NewRequestChunk()
	)
	for {
		select {
		case r := <-in:
			chunk.AddRecord(request{
				EntityKey: r.EntityKey,
				Tensor:    r.Tensor,
			})

			if chunk.RecordCount == batchSize {
				doRequest(chunk, conn)
				chunk = NewRequestChunk()
			}
		case <-ctx.Done():
			for r := range in {
				chunk.AddRecord(request{
					EntityKey: r.EntityKey,
					Tensor:    r.Tensor,
				})
			}

			if chunk.RecordCount > 0 {
				doRequest(chunk, conn)
			}
			return
		}
	}
}

type response struct {
	EntityKey         string
	InferenceResponse float64
}

type request struct {
	EntityKey string
	Tensor    []float64
}

type RequestChunk struct {
	EntityKey []string
	Tensor    []float64

	RecordCount int64
	TensorSize  int64
}

func (r *RequestChunk) AddRecord(record request) {
	r.EntityKey = append(r.EntityKey, record.EntityKey)
	r.Tensor = append(r.Tensor, record.Tensor...)

	r.RecordCount++
	r.TensorSize = int64(len(record.Tensor))
}

func (r *RequestChunk) Shape() []int64 {
	return []int64{r.RecordCount, r.TensorSize}
}

func NewRequestChunk() *RequestChunk {
	return &RequestChunk{
		EntityKey: []string{},
		Tensor:    []float64{},
	}
}

func getRequestFromFile(cancel context.CancelFunc, filePath string, records chan<- request) {
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	csvr := csv.NewReader(file)

	head, err := csvr.Read()
	if err != nil {
		panic(err)
	}
	for {
		row, err := csvr.Read()
		if err == io.EOF {
			err = nil
			return
		} else if err != nil {
			panic(err)
		}

		var (
			entityKey string
			tensor    = make([]float64, 0, len(row)-1)
		)
		for i, value := range row {
			if i == 0 {
				entityKey = value
				continue
			}
			tensor = append(tensor, mapping.GetFeatureMapping(head[i], value))
		}

		records <- request{
			EntityKey: entityKey,
			Tensor:    tensor,
		}

	}
}

func writeResponseToFile(filePath string, records <-chan response) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	count := 0
	for r := range records {
		writer.Write([]string{r.EntityKey, cast.ToString(r.InferenceResponse)})
		count++
		if count%1000 == 0 {
			log.Printf("%d record have been processed\n", count)
		}
	}

	writer.Flush()

}
