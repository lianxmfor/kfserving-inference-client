package main

import (
	"bufio"
	"context"
	"flag"
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

	lineToInference := make(chan string, worker)
	inferenceRequestToFile := make(chan float64, worker)

	go readFileByLine(inputDataPath, lineToInference)

	for i := 0; i < worker; i++ {
		go doInference(lineToInference, inferenceRequestToFile)
	}

	writeFileByLine(outputDataPath, inferenceRequestToFile)
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

func doInference(in <-chan string, out chan<- float64) {
	for s := range in {
		fp64Contents := parseRawLine(s)
		res, err := kfServingGrpcClient.Inference(context.Background(), host, &inference.ModelInferRequest{
			ModelName: modelName,
			Inputs: []*inference.ModelInferRequest_InferInputTensor{
				{
					Shape:    []int64{1, int64(len(fp64Contents))},
					Datatype: "FP64",
					Contents: &inference.InferTensorContents{
						Fp64Contents: fp64Contents,
					},
				},
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		inferenceRes := cast.ToFloat64(res.Outputs[0].Contents.Fp32Contents[0])
		out <- inferenceRes
	}

	close(out)
}

func readFileByLine(filePath string, line chan<- string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	close(line)
}

func writeFileByLine(filePath string, ch <-chan float64) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	num := 0
	for result := range ch {
		writer.WriteString(cast.ToString(result) + "\n")
		num++
		if num >= 1000 {
			writer.Flush()
			num = 0
		}
	}
	writer.Flush()
}
