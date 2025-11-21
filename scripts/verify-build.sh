#!/usr/bin/env bash
#
# verify-build.sh - Local verification script for MCP Manager
#
# Runs all quality gates that were previously enforced by GitHub Actions CI:
# - Backend unit tests with race detection
# - Backend integration tests
# - Backend contract tests
# - Go formatting check
# - Go vet static analysis
# - Staticcheck linting
# - Frontend TypeScript type checking
# - Frontend tests
# - Frontend build
# - Full Wails build (optional)
#
# Usage:
#   ./scripts/verify-build.sh              # Full verification
#   ./scripts/verify-build.sh --quick      # Fast verification (no race detection, no integration tests)
#   ./scripts/verify-build.sh --skip-build # Skip the Wails build step
#   ./scripts/verify-build.sh --skip-perf  # Skip performance benchmarks
#   ./scripts/verify-build.sh --help       # Show usage information
#
# Exit codes:
#   0  - All checks passed
#   1  - One or more checks failed
#

set -e  # Exit on first error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command line arguments
QUICK_MODE=false
SKIP_BUILD=false
SKIP_PERF=false

for arg in "$@"; do
    case $arg in
        --quick)
            QUICK_MODE=true
            ;;
        --skip-build)
            SKIP_BUILD=true
            ;;
        --skip-perf)
            SKIP_PERF=true
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --quick       Fast verification (skip race detection and integration tests)"
            echo "  --skip-build  Skip the Wails build step"
            echo "  --skip-perf   Skip performance benchmarks"
            echo "  --help        Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $arg${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Helper functions
print_step() {
    echo ""
    echo -e "${BLUE}==>${NC} ${GREEN}$1${NC}"
}

print_error() {
    echo -e "${RED}ERROR:${NC} $1"
}

print_success() {
    echo -e "${GREEN}SUCCESS:${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

# Check prerequisites
print_step "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 18 or later."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    print_error "npm is not installed. Please install npm."
    exit 1
fi

if [ "$SKIP_BUILD" = false ]; then
    if ! command -v wails &> /dev/null; then
        print_error "Wails is not installed. Please run: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        exit 1
    fi
fi

if ! command -v staticcheck &> /dev/null; then
    print_warning "staticcheck is not installed. Linting will be skipped."
    print_warning "Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
fi

echo "Prerequisites check passed."

# Download Go dependencies
print_step "Downloading Go dependencies..."
go mod download

# Run Go formatting check
print_step "Checking Go formatting..."
UNFORMATTED=$(gofmt -s -l . 2>&1 | grep -v "^vendor/" | grep -v "^frontend/" | grep -v "^build/" | grep "\.go$" || true)
if [ -n "$UNFORMATTED" ]; then
    print_error "Code is not formatted. Run 'go fmt ./...' to fix:"
    echo "$UNFORMATTED"
    exit 1
fi
print_success "Go formatting check passed."

# Run go vet
print_step "Running go vet..."
go vet ./...
print_success "go vet passed."

# Run staticcheck
if command -v staticcheck &> /dev/null; then
    print_step "Running staticcheck..."
    staticcheck ./...
    print_success "staticcheck passed."
fi

# Run backend unit tests
print_step "Running backend unit tests..."
if [ "$QUICK_MODE" = true ]; then
    go test ./internal/... -v
else
    go test ./internal/... -v -race -coverprofile=coverage.txt -covermode=atomic
fi
print_success "Backend unit tests passed."

# Run contract tests
print_step "Running contract tests..."
go test ./tests/contract/... -v
print_success "Contract tests passed."

# Run integration tests (unless in quick mode)
if [ "$QUICK_MODE" = false ]; then
    print_step "Running integration tests..."
    go test ./tests/integration/... -v -timeout=5m
    print_success "Integration tests passed."
else
    print_warning "Skipping integration tests (quick mode)."
fi

# Run performance benchmarks (optional)
if [ "$SKIP_PERF" = false ] && [ "$QUICK_MODE" = false ]; then
    print_step "Running performance benchmarks..."
    go test ./tests/performance/... -v -timeout=10m
    print_success "Performance benchmarks passed."
elif [ "$QUICK_MODE" = true ]; then
    print_warning "Skipping performance benchmarks (quick mode)."
else
    print_warning "Skipping performance benchmarks (--skip-perf flag)."
fi

# Frontend checks
print_step "Installing frontend dependencies..."
cd frontend
npm install

print_step "Running TypeScript type check..."
npm run check
print_success "TypeScript type check passed."

print_step "Running frontend tests..."
npm test
print_success "Frontend tests passed."

print_step "Building frontend..."
npm run build
print_success "Frontend build passed."

cd ..

# Build with Wails (optional)
if [ "$SKIP_BUILD" = false ]; then
    print_step "Building application with Wails..."
    wails build -clean
    print_success "Wails build passed."
else
    print_warning "Skipping Wails build (--skip-build flag)."
fi

# Final summary
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All verification checks passed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

if [ "$QUICK_MODE" = true ]; then
    print_warning "Quick mode was used. Consider running full verification before pushing."
fi

exit 0
