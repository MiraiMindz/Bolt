#!/bin/bash

set -e

# Phase 1: Setup (Same as before)
echo "--- Cleaning up and setting up directories... ---"
rm -rfv ./benchmarks/bolt_fast/bolt_fast.test ./benchmarks/bolt_sugared/bolt_sugared.test ./benchmarks/echo/echo.test ./benchmarks/gin/gin.test ./benchmarks/results/*
mkdir -pv ./benchmarks/results/{cpu,mem,trace}

# Phase 2: Run Benchmarks in Parallel with GNU Parallel
echo "--- Starting benchmarks with GNU Parallel... ---"


cd benchmarks/bolt_fast && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/bolt_fast.prof -memprofile=../results/mem/bolt_fast.prof -trace=../results/trace/bolt_fast.trace . > ../results/bolt_fast.json
cd ../..
cd benchmarks/bolt_sugared && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/bolt_sugared.prof -memprofile=../results/mem/bolt_sugared.prof -trace=../results/trace/bolt_sugared.trace . > ../results/bolt_sugared.json
cd ../..
cd benchmarks/gin && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/gin.prof -memprofile=../results/mem/gin.prof -trace=../results/trace/gin.trace . > ../results/gin.json
cd ../..
cd benchmarks/echo && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/echo.prof -memprofile=../results/mem/echo.prof -trace=../results/trace/echo.trace . > ../results/echo.json
cd ../..

# Phase 3: Analyze Profiles in Parallel with GNU Parallel
echo "--- Analyzing profiles with GNU Parallel... ---"


go tool pprof -text ./benchmarks/results/cpu/bolt_sugared.prof > ./benchmarks/results/cpu/bolt_sugared_cpu.txt
go tool pprof -text ./benchmarks/results/cpu/bolt_fast.prof > ./benchmarks/results/cpu/bolt_fast_cpu.txt
go tool pprof -text ./benchmarks/results/cpu/gin.prof > ./benchmarks/results/cpu/gin_cpu.txt
go tool pprof -text ./benchmarks/results/cpu/echo.prof > ./benchmarks/results/cpu/echo_cpu.txt
go tool pprof -text ./benchmarks/results/mem/bolt_sugared.prof > ./benchmarks/results/mem/bolt_sugared_mem.txt
go tool pprof -text ./benchmarks/results/mem/bolt_fast.prof > ./benchmarks/results/mem/bolt_fast_mem.txt
go tool pprof -text ./benchmarks/results/mem/gin.prof > ./benchmarks/results/mem/gin_mem.txt
go tool pprof -text ./benchmarks/results/mem/echo.prof > ./benchmarks/results/mem/echo_mem.txt

# Phase 4: Final Summary (Same as before)
echo "--- Generating final benchstat summary... ---"
benchstat -table \
  <(cat ./benchmarks/results/bolt_sugared.json | go run ./internal/tools/json2bench/json2bench.go) \
  <(cat ./benchmarks/results/bolt_fast.json | go run ./internal/tools/json2bench/json2bench.go) \
  <(cat ./benchmarks/results/gin.json | go run ./internal/tools/json2bench/json2bench.go) \
  <(cat ./benchmarks/results/echo.json | go run ./internal/tools/json2bench/json2bench.go) \
  > ./benchmarks/results/summary.txt

echo "--- All done! ---"