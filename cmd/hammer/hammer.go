// Binary hammer sends requests to your Raft cluster as fast as it can.
// It sends the written out version of the Dutch numbers up to 2000.
// In the end it asks the Raft cluster what the longest three words were.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/Jille/grpc-multi-resolver"
	pb "github.com/Jille/raft-grpc-example/proto"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	ntw "moul.io/number-to-words"
)

func main() {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "Example"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	conn, err := grpc.Dial("multi:///localhost:50051,localhost:50052,localhost:50053",
		grpc.WithDefaultServiceConfig(serviceConfig), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		log.Fatalf("dialing failed: %v", err)
	}
	defer conn.Close()
	c := pb.NewExampleClient(conn)

	ch := generateWords()

	var wg sync.WaitGroup
	for i := 0; 10 > i; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range ch {
				_, err := c.AddWord(context.Background(), &pb.AddWordRequest{Word: w})
				if err != nil {
					log.Fatalf("AddWord RPC failed: %v", err)
				}
			}
		}()
	}
	wg.Wait()
	resp, err := c.GetWords(context.Background(), &pb.GetWordsRequest{})
	if err != nil {
		log.Fatalf("GetWords RPC failed: %v", err)
	}
	fmt.Println(resp)
}

func generateWords() <-chan string {
	ch := make(chan string, 1)
	go func() {
		for i := 1; 2000 > i; i++ {
			ch <- ntw.IntegerToNlNl(i)
		}
		close(ch)
	}()
	return ch
}
