package workerpool

import (
	"context"
	"errors"
	"sync"
)

type Job[T any] struct {
	ID int
	Fn func(context.Context) (T, error)
}

// WorkerPool is a simple worker pool that can be used to process jobs concurrently.
func WorkerPool[T any](ctx context.Context, totalWorkers int, jobsList []Job[T], handleResult func(T)) error {

	if len(jobsList) == 0 {
		return errors.New("Jobs list is empty!")
	}

	if totalWorkers <= 0 {
		return errors.New("total workers must be greater than zero")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workerCount := min(totalWorkers, len(jobsList))

	jobs := make(chan Job[T], len(jobsList))
	results := make(chan T, len(jobsList))
	errors := make(chan error, 1)

	var wg sync.WaitGroup

	// workers
	for _ = range workerCount {

		wg.Go(func() {
			worker(ctx, jobs, results, errors)
		})
	}

	// jobs producer
	for _, job := range jobsList {
		jobs <- job
	}
	close(jobs)

	// close channels after workers finish
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err, ok := <-errors:
			if ok && err != nil {
				cancel()
				return err
			}

		case result, ok := <-results:
			if !ok {
				return nil
			}
			handleResult(result)
		}
	}
}

func worker[T any](ctx context.Context, jobs <-chan Job[T], results chan<- T, errors chan<- error) {

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			result, err := job.Fn(ctx)

			if err != nil {
				select {
				case errors <- err:
				default:
				}
				return
			}

			select {
			case <-ctx.Done():
				return
			case results <- result:
			}
		}
	}
}
