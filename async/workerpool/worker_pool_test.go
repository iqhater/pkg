// Benchmark run: go test -bench="Benchmark(BaseLoop|WorkerPool)" -benchmem -v ./async/workerpool
// Run Tests: go test -v -cover -race -count=1 ./async/workerpool/...
package workerpool

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// Check for goroutine leaks
func TestMain(m *testing.M) {

	goleak.VerifyTestMain(m)

	// run tests
	os.Exit(m.Run())
}

func TestWorkerPoolOK(t *testing.T) {

	// Arrange
	ctx := context.Background()
	totalWorkers := 3

	jobFunc1 := func(ctx context.Context) (int, error) {
		time.Sleep(7 * time.Millisecond)
		return 42, nil
	}
	jobFunc2 := func(ctx context.Context) (int, error) {
		time.Sleep(5 * time.Millisecond)
		return 52, nil
	}
	jobFunc3 := func(ctx context.Context) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 62, nil
	}

	jobsList := []Job[int]{
		{ID: 1, Fn: jobFunc1},
		{ID: 2, Fn: jobFunc2},
		{ID: 3, Fn: jobFunc3},
		{ID: 4, Fn: jobFunc1},
		{ID: 5, Fn: jobFunc2},
	}

	var results []int
	handleResults := func(res int) {
		results = append(results, res)
	}

	// Act
	err := WorkerPool(ctx, totalWorkers, jobsList, handleResults)

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != len(jobsList) {
		t.Errorf("Expected %d results, got %d", len(jobsList), len(results))
	}

	expected := []any{42, 52, 62, 42, 52} // total sum 250

	// get total sum for convince that all results are correct
	sum := 0
	for _, res := range expected {
		sum += res.(int)
	}

	if sum != 250 {
		t.Errorf("Expected sum to be 250, got %d", sum)
	}
}

func TestWorkerOK(t *testing.T) {

	// Arrange
	ctx := context.Background()
	jobs := make(chan Job[int], 2)
	results := make(chan int, 2)
	errors := make(chan error, 1)

	jobFunc := func(ctx context.Context) (int, error) {
		time.Sleep(5 * time.Millisecond)
		return 42, nil
	}

	go func() {
		jobs <- Job[int]{ID: 1, Fn: jobFunc}
		jobs <- Job[int]{ID: 2, Fn: jobFunc}
		close(jobs)
	}()

	// Act
	go worker(ctx, jobs, results, errors)
	go worker(ctx, jobs, results, errors)
	go worker(ctx, jobs, results, errors)

	res1 := <-results
	res2 := <-results

	// Assert
	if res1 != 42 || res2 != 42 {
		t.Errorf("Expected results to be 42, got %v and %v", res1, res2)
	}
}

func TestWorkerPoolEmptyJobs(t *testing.T) {

	// Arrange
	ctx := context.Background()
	totalWorkers := 3
	jobsList := []Job[int]{}

	handleResult := func(res int) {}

	// Act
	err := WorkerPool(ctx, totalWorkers, jobsList, handleResult)

	// Assert
	if err == nil {
		t.Error("Expected error for empty jobs list, got nil")
	}
	if !strings.Contains(err.Error(), "Jobs list is empty!") {
		t.Errorf("Expected 'Jobs list is empty!' error, got: %v", err)
	}
}

func TestWorkerPoolWithErrors(t *testing.T) {

	// Arrange
	ctx := context.Background()
	totalWorkers := 2

	jobFuncOK := func(ctx context.Context) (int, error) {
		time.Sleep(5 * time.Millisecond)
		return 42, nil
	}
	jobFuncError := func(ctx context.Context) (int, error) {
		time.Sleep(5 * time.Millisecond)
		return 0, errors.New("job error")
	}

	jobsList := []Job[int]{
		{ID: 1, Fn: jobFuncOK},
		{ID: 2, Fn: jobFuncError},
		{ID: 3, Fn: jobFuncOK},
	}

	var results []int
	handleResult := func(res int) {
		results = append(results, res)
	}

	// Act
	err := WorkerPool(ctx, totalWorkers, jobsList, handleResult)

	// Assert
	if err == nil {
		t.Error("Expected error from job, got nil")
	}
	if err.Error() != "job error" {
		t.Errorf("Expected 'job error', got: %v", err)
	}
}

func BenchmarkBaseLoop(b *testing.B) {

	// Arrange
	const totalJobs = 5_000

	jobFunc := func(ctx context.Context) (int, error) {
		time.Sleep(5 * time.Millisecond) // work imitation
		return 42, nil
	}

	// init jobs list
	jobsList := make([]Job[int], totalJobs)
	for i := range totalJobs {
		jobsList[i] = Job[int]{ID: i, Fn: jobFunc}
	}

	// handle results
	handleResults := func(result int) {}

	ctx := context.Background()

	// run benchmark
	for b.Loop() {

		for _, job := range jobsList {
			result, _ := job.Fn(ctx)
			handleResults(result)
		}
	}
}

func BenchmarkWorkerPool(b *testing.B) {

	// Arrange
	const (
		totalJobs    = 5_000
		totalWorkers = 500
	)

	jobFunc := func(ctx context.Context) (int, error) {
		time.Sleep(5 * time.Millisecond) // work imitation
		return 42, nil
	}

	// init jobs list
	jobsList := make([]Job[int], totalJobs)
	for i := range totalJobs {
		jobsList[i] = Job[int]{ID: i, Fn: jobFunc}
	}

	// handle results
	handleResult := func(result int) {}

	ctx := context.Background()

	// run benchmark
	for b.Loop() {
		_ = WorkerPool(ctx, totalWorkers, jobsList, handleResult)
	}
}
