package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"kfserving-inference-client/inference"

	"github.com/spf13/cast"
)

var (
	inputDataPath  string
	outputDataPath string
	host           string
	modelName      string
	worker         int
)

func init() {
	InitKFServingGrpcClient(time.Second * 5)

	flag.StringVar(&inputDataPath, "i", "", "The local filestore path where the input file with the data to process is located")
	flag.StringVar(&outputDataPath, "o", "", "The local filestore path where the output file should be written with the outputs of the batch processing")
	flag.StringVar(&host, "host", "", "The hostname for the seldon model to send the request to, which can be the ingress of the Seldon model or the service itself ")
	flag.StringVar(&modelName, "m", "", "model name")
	flag.IntVar(&worker, "w", 1, "The number of parallel request processor workers to run for parallel processing")
}

func main() {
	flag.Parse()

	requestRecords := make(chan RequestRecord, worker)
	responseRecords := make(chan ResponseRecord, worker)

	go readFileByLine(inputDataPath, requestRecords)

	for i := 0; i < worker; i++ {
		go doInference(requestRecords, responseRecords)
	}

	writeFileByLine(outputDataPath, responseRecords)
}

func parseRawLine(s string) []float64 {
	s = strings.TrimFunc(s, func(r rune) bool {
		return r == '{' || r == '}'
	})
	ss := strings.Split(strings.TrimSpace(s), ",")

	fp64Contents := make([]float64, len(ss))
	for i, v := range ss {
		fp64Contents[i] = cast.ToFloat64(v)
	}
	return fp64Contents
}

func doInference(records <-chan RequestRecord, out chan<- ResponseRecord) {
	for r := range records {
		res, err := kfServingGrpcClient.Inference(context.Background(), host, &inference.ModelInferRequest{
			ModelName: modelName,
			Inputs: []*inference.ModelInferRequest_InferInputTensor{
				{
					Shape:    []int64{1, int64(len(r.Tensor))},
					Datatype: "FP64",
					Contents: &inference.InferTensorContents{
						Fp64Contents: r.Tensor,
					},
				},
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		out <- ResponseRecord{
			EntityKey:         r.EntityKey,
			InferenceResponse: cast.ToFloat64(res.Outputs[0].Contents.Fp32Contents[0]),
		}
	}

	close(out)
}

type RequestRecord struct {
	EntityKey string
	Tensor    []float64
}

type ResponseRecord struct {
	EntityKey         string
	InferenceResponse float64
}

func readFileByLine(filePath string, records chan<- RequestRecord) {
	defer close(records)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	csvr := csv.NewReader(file)

	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		entityKey := row[0]
		tensor := make([]float64, 0, len(row)-1)
		for _, t := range row[1:] {
			tensor = append(tensor, cast.ToFloat64(t))
		}

		records <- RequestRecord{
			EntityKey: entityKey,
			Tensor:    tensor,
		}

	}
}

func writeFileByLine(filePath string, records <-chan ResponseRecord) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	num := 0
	total := 0
	for r := range records {
		writer.Write([]string{r.EntityKey, cast.ToString(r.InferenceResponse)})
		num++
		if num >= 1000 {
			writer.Flush()
			num = 0
		}
		total++
		if total%10 == 0 {
			fmt.Printf("%d record have been processed", total)
		}
	}

	writer.Flush()
}
