#!/bin/bash
# This version requires GNU Parallel to be installed (e.g., `sudo apt-get install parallel`)

set -e

# Phase 1: Setup (Same as before)
echo "--- Cleaning up and setting up directories... ---"
rm -rfv ./benchmarks/bolt_fast/bolt_fast.test ./benchmarks/bolt_sugared/bolt_sugared.test ./benchmarks/echo/echo.test ./benchmarks/gin/gin.test ./benchmarks/results/*
mkdir -pv ./benchmarks/results/{cpu,mem,trace}

# Phase 2: Run Benchmarks in Parallel with GNU Parallel
echo "--- Starting benchmarks with GNU Parallel... ---"
parallel --jobs 3 --bar --tag <<-'EOF'
  cd benchmarks/bolt_fast && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/bolt_fast.prof -memprofile=../results/mem/bolt_fast.prof -trace=../results/trace/bolt_fast.trace . > ../results/bolt_fast.json
  cd benchmarks/bolt_sugared && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/bolt_sugared.prof -memprofile=../results/mem/bolt_sugared.prof -trace=../results/trace/bolt_sugared.trace . > ../results/bolt_sugared.json
  cd benchmarks/gin && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/gin.prof -memprofile=../results/mem/gin.prof -trace=../results/trace/gin.trace . > ../results/gin.json
  cd benchmarks/echo && go test -bench=. -count=8 -json -benchmem -cpuprofile=../results/cpu/echo.prof -memprofile=../results/mem/echo.prof -trace=../results/trace/echo.trace . > ../results/echo.json
EOF

# Phase 3: Analyze Profiles in Parallel with GNU Parallel
echo "--- Analyzing profiles with GNU Parallel... ---"
parallel --jobs 0 --tag <<-'EOF'
  go tool pprof -text ./benchmarks/results/cpu/bolt_sugared.prof > ./benchmarks/results/cpu/bolt_sugared_cpu.txt
  go tool pprof -text ./benchmarks/results/cpu/bolt_fast.prof > ./benchmarks/results/cpu/bolt_fast_cpu.txt
  go tool pprof -text ./benchmarks/results/cpu/gin.prof > ./benchmarks/results/cpu/gin_cpu.txt
  go tool pprof -text ./benchmarks/results/cpu/echo.prof > ./benchmarks/results/cpu/echo_cpu.txt
  go tool pprof -text ./benchmarks/results/mem/bolt_sugared.prof > ./benchmarks/results/mem/bolt_sugared_mem.txt
  go tool pprof -text ./benchmarks/results/mem/bolt_fast.prof > ./benchmarks/results/mem/bolt_fast_mem.txt
  go tool pprof -text ./benchmarks/results/mem/gin.prof > ./benchmarks/results/mem/gin_mem.txt
  go tool pprof -text ./benchmarks/results/mem/echo.prof > ./benchmarks/results/mem/echo_mem.txt
EOF

# Phase 4: Final Summary (Same as before)
echo "--- Generating final benchstat summary... ---"
benchstat -table \
    <(cat ./benchmarks/results/bolt_sugared.json | go run ./internal/tools/json2bench/json2bench.go) \
    <(cat ./benchmarks/results/bolt_fast.json | go run ./internal/tools/json2bench/json2bench.go) \
    <(cat ./benchmarks/results/gin.json | go run ./internal/tools/json2bench/json2bench.go) \
    <(cat ./benchmarks/results/echo.json | go run ./internal/tools/json2bench/json2bench.go) \
    > ./benchmarks/results/summary.txt

echo "--- All done! ---"

# #!/usr/bin/env bash

# # Exit immediately if a command exits with a non-zero status.
# # Exit on unset variables, and propagate exit status through pipes.
# set -euo pipefail

# # --- Configuration ---
# FRAMEWORKS=("bolt" "gin" "echo")
# BENCH_DIR="./benchmarks"
# RESULTS_DIR="${BENCH_DIR}/results"
# BENCH_COUNT=8

# # --- Script Start ---

# echo "--- Step 1: Cleaning up old results and test binaries ---"
# # Use a loop to remove test binaries, ignoring errors if they don't exist
# for framework in "${FRAMEWORKS[@]}"; do
#     rm -f "${BENCH_DIR}/${framework}/${framework}.test"
# done
# # Remove the entire old results directory
# rm -rf "${RESULTS_DIR}"
# echo "Cleanup complete."
# echo

# echo "--- Step 2: Creating results directory structure ---"
# # -p creates parent directories if needed, and doesn't error if it already exists.
# mkdir -p "${RESULTS_DIR}"/{cpu,mem,trace}
# echo "Directory structure created."
# echo

# echo "--- Step 3: Running benchmarks for each framework ---"
# for framework in "${FRAMEWORKS[@]}"; do
#     echo "Benchmarking ${framework}..."
    
#     # Run the test in a subshell to avoid manually changing directories
#     (
#         cd "${BENCH_DIR}/${framework}" || exit 1
        
#         go test \
#             -bench=. \
#             -count=${BENCH_COUNT} \
#             -json \
#             -benchmem \
#             -cpuprofile="../results/cpu/${framework}.prof" \
#             -memprofile="../results/mem/${framework}.prof" \
#             -trace="../results/trace/${framework}.trace" \
#             . > "../results/${framework}.json"
#     )
    
#     echo "Finished benchmarking ${framework}."
# done
# echo "All benchmarks completed."
# echo

# echo "--- Step 4: Processing profiling data into text summaries ---"
# for framework in "${FRAMEWORKS[@]}"; do
#     echo "Processing profiles for ${framework}..."
    
#     # Process CPU profile
#     go tool pprof -text "${RESULTS_DIR}/cpu/${framework}.prof" > "${RESULTS_DIR}/cpu/${framework}_cpu.txt"
    
#     # Process Memory profile
#     go tool pprof -text "${RESULTS_DIR}/mem/${framework}.prof" > "${RESULTS_DIR}/mem/${framework}_mem.txt"
# done
# echo "All profiles processed."
# echo

# echo "--- Step 5: Generating final summary report with benchstat ---"
# benchstat -table \
#     <(cat "${RESULTS_DIR}/bolt.json"  | go run ./internal/tools/json2bench/json2bench.go) \
#     <(cat "${RESULTS_DIR}/gin.json"   | go run ./internal/tools/json2bench/json2bench.go) \
#     <(cat "${RESULTS_DIR}/echo.json"  | go run ./internal/tools/json2

