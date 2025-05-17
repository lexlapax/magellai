# Dependency and Binary Size Reduction Analysis

This document analyzes opportunities for reducing dependencies and binary size in the Magellai project.

## Current State

### Binary Size
- Current binary: 14M
- With optimization flags (`-ldflags="-s -w"`): 10M
- Potential with all optimizations: 3.5-4.5M

### Dependency Count
- Started with: 50 dependencies
- After go-llms update: 42 dependencies
- After pflag removal: 40 dependencies
- Potential with optimization: ~35-37 dependencies

## Optimization Analysis

### 1. Build-Time Optimizations (No Code Changes)

#### Quick Wins
1. **Compiler Flags**
   ```bash
   go build -ldflags="-s -w" -trimpath -o magellai cmd/magellai/main.go
   ```
   - `-s`: Strip symbol table
   - `-w`: Strip DWARF debug info
   - `-trimpath`: Remove file system paths
   - **Result**: 14M → 10M (28% reduction)

2. **UPX Compression**
   ```bash
   upx --best magellai
   ```
   - Compresses the binary
   - **Result**: 10M → 4-5M (additional 50-60% reduction)
   - **Trade-off**: Slightly slower startup time

### 2. Dependency Analysis

#### Largest Dependencies

1. **go-llms (github.com/lexlapax/go-llms)**
   - Core LLM functionality
   - **Cannot be replaced**
   - Includes: providers, agents, tools, workflows

2. **koanf (github.com/knadh/koanf/v2)**
   - Configuration management
   - Multiple providers included
   - **Optimization potential**: Medium

3. **kong (github.com/alecthomas/kong)**
   - CLI parsing framework
   - Relatively lightweight
   - **Replacement effort**: High

4. **kongplete (github.com/willabides/kongplete)**
   - Shell completion support
   - **Can be removed**: Easy

5. **json-iterator/go**
   - High-performance JSON parser
   - Used by go-llms (not directly)
   - **Cannot replace**: Part of go-llms

### 3. Koanf Provider Analysis

#### Currently Used Providers

1. **Required Providers** ✅
   - `file`: Loading YAML configuration files
   - `env`: Environment variables with `MAGELLAI_` prefix
   - `yaml`: YAML parser for config files

2. **Optional Providers** ⚠️
   - `confmap`: Used for defaults and command-line overrides
   - `rawbytes`: Used in one utility function only

3. **Unused Providers** ❌
   - `posflag`: Already removed with pflag

#### Actual Usage
```go
// Required imports
"github.com/knadh/koanf/v2"
"github.com/knadh/koanf/parsers/yaml"
"github.com/knadh/koanf/providers/env"
"github.com/knadh/koanf/providers/file"

// Optional (can be removed)
"github.com/knadh/koanf/providers/confmap"
"github.com/knadh/koanf/providers/rawbytes"
```

## Recommended Implementation Plan

### Phase 1: No Code Changes (Immediate)
1. Update build process:
   ```makefile
   build-optimized:
       go build -ldflags="-s -w -X main.version=$(VERSION)" \
                -trimpath \
                -o magellai \
                cmd/magellai/main.go
       upx --best magellai
   ```
2. **Expected reduction**: 14M → 4-5M (70% reduction)

### Phase 2: Easy Code Changes
1. **Remove kongplete**
   - Remove shell completion functionality
   - Delete imports and InstallCompletions command
   - **Savings**: ~200KB

2. **Remove unused koanf providers**
   - Remove confmap and rawbytes providers
   - Update configuration loading code
   - **Savings**: ~50-100KB

### Phase 3: Medium Effort Changes
1. **Simplify koanf usage**
   - Use only file + env providers
   - Move defaults to a YAML file
   - **Savings**: ~100-200KB

## Impact Summary

### Binary Size Reduction Path
1. Current: **14M**
2. With compiler flags: **10M** (-28%)
3. With UPX compression: **4-5M** (-71%)
4. With dependency cleanup: **3.5-4.5M** (-75%)

### Dependency Reduction Path
1. Current: **40 dependencies**
2. Remove kongplete: **39 dependencies**
3. Remove unused koanf providers: **37 dependencies**
4. Final count: **~35-37 dependencies**

## Not Recommended

### High-Effort, Low-Reward Changes
1. **Replacing Kong CLI**
   - Would require complete CLI restructure
   - Kong is already lightweight
   - Not worth the effort

2. **Replacing go-llms**
   - Core functionality
   - Cannot be replaced
   - Would require complete rewrite

3. **Replacing json-iterator**
   - Part of go-llms, not directly used
   - No direct benefit to magellai

## Implementation Guide

### Step 1: Update Makefile
```makefile
# Add optimized build target
build-optimized:
	go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" \
	         -trimpath \
	         -o bin/magellai \
	         cmd/magellai/main.go

# Add compressed build target
build-compressed: build-optimized
	upx --best bin/magellai
```

### Step 2: Remove Kongplete (Optional)
```go
// Remove from main.go
// import "github.com/willabides/kongplete"

// Remove from CLI struct
// InstallCompletions kongplete.InstallCompletions `cmd:"" help:"Install shell completions" group:"config"`
```

### Step 3: Clean Koanf Providers (Optional)
1. Remove unused imports from `pkg/config/config.go`
2. Update configuration loading to not use confmap
3. Remove or simplify `LoadFromRawBytes` function

## Conclusion

The most impactful optimization is using proper build flags and UPX compression, achieving **75% size reduction** with no code changes. Additional dependency cleanup provides marginal benefits but may improve maintainability.

### Recommended Approach
1. Implement build optimizations immediately (no code changes)
2. Consider removing kongplete if shell completion isn't critical
3. Clean up koanf providers during next refactoring

### Final Metrics
- **Binary size**: 14M → 3.5-4.5M (75% reduction)
- **Dependencies**: 40 → 35-37 (7-12% reduction)
- **Code changes**: Minimal to none