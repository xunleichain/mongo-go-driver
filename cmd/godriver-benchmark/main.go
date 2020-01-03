package main

import (
	"os"

	"go.mongodb.org/mongo-driver/benchmark"
)

func main() {
	os.Exit(benchmark.DriverBenchmarkMain())
}
