package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/iqhater/pkg/async/workerpool"
)

// WorkerPool example
func main() {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	jobs := []workerpool.Job[int]{
		{
			ID: 1,
			Fn: func(ctx context.Context) (int, error) {
				time.Sleep(time.Duration(r.Intn(5)) * time.Second)
				return 6 * r.Intn(8), nil
			},
		},
		{
			ID: 2,
			Fn: func(ctx context.Context) (int, error) {
				time.Sleep(time.Duration(r.Intn(3)) * time.Second)
				return 2 * r.Intn(3), nil
			},
		},
		{
			ID: 3,
			Fn: func(ctx context.Context) (int, error) {
				time.Sleep(time.Duration(r.Intn(2)) * time.Second)
				return -1, errors.New("job 3 failed!") // if job failed, then stop others jobs
				// return 5 * r.Intn(4), nil
			},
		},
		{
			ID: 4,
			Fn: func(ctx context.Context) (int, error) {
				time.Sleep(time.Duration(r.Intn(4)) * time.Second)
				return 3 * r.Intn(2), nil
			},
		},
		{
			ID: 5,
			Fn: func(ctx context.Context) (int, error) {
				time.Sleep(time.Duration(r.Intn(3)) * time.Second)
				return 1 * r.Intn(3), nil
			},
		},
	}

	const totalWorkers = 3

	handleResult := func(result int) {
		fmt.Println(result)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	t := time.Now()
	err := workerpool.WorkerPool(
		ctx,
		totalWorkers,
		jobs,
		handleResult,
	)

	if err != nil {
		fmt.Println("worker pool failed:", err)
	}
	fmt.Printf("total time: %v\n", time.Since(t))
}
