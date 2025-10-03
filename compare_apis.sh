#!/bin/bash

# 🔥 Bolt Sugared vs Fast API Comparison Tool
# ============================================

set -e

echo "🚀 BOLT API COMPARISON SUITE"
echo "============================"
echo ""
echo "Comparing three implementations:"
echo "  1. 🍬 Sugared API (ergonomic, map-based)"
echo "  2. ⚡ Fast API (zero-allocation, field-based)"
echo "  3. 🎯 Gin (competitor framework)"
echo "  4. 🌊 Echo (competitor framework)"
echo ""

# Create results directory
mkdir -p benchmarks/results/comparison

echo "📊 Running benchmarks (this may take a few minutes)..."
echo ""

# Function to run benchmarks and save results
run_benchmark() {
    local name=$1
    local package=$2
    local pattern=$3
    local output=$4

    echo "  ⏳ Running $name benchmarks..."
    cd "$package"
    go test -bench="$pattern" -benchmem -count=5 -run=^$ > "../../results/comparison/$output" 2>&1
    cd - > /dev/null
    echo "  ✅ $name complete"
}

# Run Bolt Sugared benchmarks
run_benchmark "Bolt Sugared" "benchmarks/bolt" "BenchmarkSugared" "bolt_sugared.txt"

# Run Bolt Fast benchmarks
run_benchmark "Bolt Fast" "benchmarks/bolt" "BenchmarkFast" "bolt_fast.txt"

# Run Gin benchmarks
run_benchmark "Gin" "benchmarks/gin" "Benchmark" "gin.txt"

# Run Echo benchmarks
run_benchmark "Echo" "benchmarks/echo" "Benchmark" "echo.txt"

echo ""
echo "📈 Generating comparison report..."

# Generate comparison report
cat > benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'
# 🔥 Bolt API Performance Comparison

## Test Environment
- **Date:** $(date)
- **Go Version:** $(go version)
- **CPU:** $(lscpu | grep "Model name" | cut -d: -f2 | xargs)

## Benchmarks Compared

### 1. 🍬 **Bolt Sugared API**
- Uses `Context` with `map[string]interface{}`
- Ergonomic, developer-friendly
- Small allocation overhead

### 2. ⚡ **Bolt Fast API**
- Uses `FastContext` with strongly-typed fields
- Zero-allocation where possible
- Maximum performance

### 3. 🎯 **Gin Framework**
- Popular Go web framework
- Industry standard

### 4. 🌊 **Echo Framework**
- Popular Go web framework
- Known for performance

---

## Results Summary

### Static Route Performance

**Bolt Sugared:**
EOF

# Extract and format results
echo "$(grep 'BenchmarkSugaredStaticRoute' benchmarks/results/comparison/bolt_sugared.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Bolt Fast:**
EOF

echo "$(grep 'BenchmarkFastStaticRoute' benchmarks/results/comparison/bolt_fast.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Gin:**
EOF

echo "$(grep 'BenchmarkStaticRoute' benchmarks/results/comparison/gin.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Echo:**
EOF

echo "$(grep 'BenchmarkStaticRoute' benchmarks/results/comparison/echo.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

---

### Dynamic Route Performance

**Bolt Sugared:**
EOF

echo "$(grep 'BenchmarkSugaredDynamicRoute' benchmarks/results/comparison/bolt_sugared.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Bolt Fast:**
EOF

echo "$(grep 'BenchmarkFastDynamicRoute' benchmarks/results/comparison/bolt_fast.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Gin:**
EOF

echo "$(grep 'BenchmarkDynamicRoute' benchmarks/results/comparison/gin.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Echo:**
EOF

echo "$(grep 'BenchmarkDynamicRoute' benchmarks/results/comparison/echo.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

---

### Typed JSON Performance

**Bolt Sugared:**
EOF

echo "$(grep 'BenchmarkSugaredTypedJSON' benchmarks/results/comparison/bolt_sugared.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Bolt Fast:**
EOF

echo "$(grep 'BenchmarkFastTypedJSON' benchmarks/results/comparison/bolt_fast.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Gin:**
EOF

echo "$(grep 'BenchmarkTypedJSON' benchmarks/results/comparison/gin.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Echo:**
EOF

echo "$(grep 'BenchmarkTypedJSON' benchmarks/results/comparison/echo.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

---

### Middleware Performance

**Bolt Sugared:**
EOF

echo "$(grep 'BenchmarkSugaredMiddleware' benchmarks/results/comparison/bolt_sugared.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Bolt Fast:**
EOF

echo "$(grep 'BenchmarkFastMiddleware' benchmarks/results/comparison/bolt_fast.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Gin:**
EOF

echo "$(grep 'BenchmarkMiddleware' benchmarks/results/comparison/gin.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

**Echo:**
EOF

echo "$(grep 'BenchmarkMiddleware' benchmarks/results/comparison/echo.txt | tail -1)" >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'

---

## Key Findings

### ⚡ Fast API Advantages
- **Fewer Allocations:** Strongly-typed fields eliminate intermediate map allocations
- **No Reflection:** Direct field encoding skips reflection overhead
- **Type Safety:** Compile-time type checking prevents runtime errors
- **Predictable Performance:** Zero-allocation guarantees consistent latency

### 🍬 Sugared API Advantages
- **Developer Ergonomics:** Natural Go syntax with maps
- **Rapid Prototyping:** Quick to write and iterate
- **Flexibility:** Easy to add/remove fields dynamically
- **Still Fast:** Competitive performance for most use cases

### 📊 When to Use Which

**Use Fast API when:**
- ⚡ High-throughput services (>10K req/sec)
- ⚡ Strict latency requirements
- ⚡ Every allocation matters
- ⚡ Production hot paths

**Use Sugared API when:**
- 🍬 Prototyping and development
- 🍬 Internal tools and dashboards
- 🍬 Moderate traffic (<1K req/sec)
- 🍬 Developer productivity > microseconds

---

## Full Results

### Bolt Sugared API
```
EOF

cat benchmarks/results/comparison/bolt_sugared.txt >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'
```

### Bolt Fast API
```
EOF

cat benchmarks/results/comparison/bolt_fast.txt >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'
```

### Gin Framework
```
EOF

cat benchmarks/results/comparison/gin.txt >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'
```

### Echo Framework
```
EOF

cat benchmarks/results/comparison/echo.txt >> benchmarks/results/comparison/COMPARISON_REPORT.md

cat >> benchmarks/results/comparison/COMPARISON_REPORT.md << 'EOF'
```

---

## Conclusion

Bolt provides **BOTH APIs** - giving you the power to choose:
- **Sugared** for ergonomics and rapid development
- **Fast** for maximum performance and zero allocations

**Choose the right tool for each endpoint!** 🚀
EOF

echo ""
echo "✅ COMPARISON COMPLETE!"
echo ""
echo "📄 Results saved to: benchmarks/results/comparison/COMPARISON_REPORT.md"
echo ""
echo "🔍 Quick Summary:"
echo ""

# Show quick comparison of static routes
echo "Static Route Comparison:"
echo "  Bolt Sugared: $(grep 'BenchmarkSugaredStaticRoute' benchmarks/results/comparison/bolt_sugared.txt | tail -1 | awk '{print $3, $4, $5, $6}')"
echo "  Bolt Fast:    $(grep 'BenchmarkFastStaticRoute' benchmarks/results/comparison/bolt_fast.txt | tail -1 | awk '{print $3, $4, $5, $6}')"
echo "  Gin:          $(grep 'BenchmarkStaticRoute' benchmarks/results/comparison/gin.txt | tail -1 | awk '{print $3, $4, $5, $6}')"
echo "  Echo:         $(grep 'BenchmarkStaticRoute' benchmarks/results/comparison/echo.txt | tail -1 | awk '{print $3, $4, $5, $6}')"
echo ""

echo "📖 View full report: cat benchmarks/results/comparison/COMPARISON_REPORT.md"
echo ""
echo "🎉 Done! Bolt gives you the power to choose your performance level! ⚡"
