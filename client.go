package main

import (
	"context"
	"log"
	"sync"
	"time"

	"kfserving-inference-client/inference"

	grpc "google.golang.org/grpc"
)

var kfServingGrpcClient *KFServingGrpcClient

type KFServingGrpcClient struct {
	mutex           sync.Mutex
	dialCallOptions []grpc.DialOption
	conns           map[string]*grpc.ClientConn
}

func InitKFServingGrpcClient(d time.Duration) {
	kfServingGrpcClient = NewKFServingGrpcClient(grpc.WithInsecure(), grpc.WithTimeout(d))
}

func NewKFServingGrpcClient(callOptions ...grpc.DialOption) *KFServingGrpcClient {
	return &KFServingGrpcClient{
		dialCallOptions: callOptions,
		conns:           make(map[string]*grpc.ClientConn),
	}
}

func (k *KFServingGrpcClient) getConnection(host string) (*grpc.ClientConn, error) {
	k.mutex.Lock()
	if _, ok := k.conns[host]; !ok {
		log.Printf("dial host %s", host)
		if conn, err := grpc.Dial(host, k.dialCallOptions...); err != nil {
			return nil, err
		} else {
			k.conns[host] = conn
		}
	}
	k.mutex.Unlock()
	return k.conns[host], nil
}

func (k *KFServingGrpcClient) Inference(ctx context.Context, host string, r *inference.ModelInferRequest) (*inference.ModelInferResponse, error) {
	conn, err := k.getConnection(host)
	if err != nil {
		return nil, err
	}
	grpcClient := inference.NewGRPCInferenceServiceClient(conn)

	return grpcClient.ModelInfer(ctx, r)
}
