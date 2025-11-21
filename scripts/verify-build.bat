@echo off
REM verify-build.bat - Local verification script for MCP Manager (Windows)
REM
REM Runs all quality gates that were previously enforced by GitHub Actions CI:
REM - Backend unit tests with race detection
REM - Backend integration tests
REM - Backend contract tests
REM - Go formatting check
REM - Go vet static analysis
REM - Staticcheck linting
REM - Frontend TypeScript type checking
REM - Frontend tests
REM - Frontend build
REM - Full Wails build (optional)
REM
REM Usage:
REM   scripts\verify-build.bat              # Full verification
REM   scripts\verify-build.bat --quick      # Fast verification (no race detection, no integration tests)
REM   scripts\verify-build.bat --skip-build # Skip the Wails build step
REM   scripts\verify-build.bat --skip-perf  # Skip performance benchmarks
REM   scripts\verify-build.bat --help       # Show usage information
REM
REM Exit codes:
REM   0  - All checks passed
REM   1  - One or more checks failed
REM

setlocal enabledelayedexpansion

REM Parse command line arguments
set QUICK_MODE=false
set SKIP_BUILD=false
set SKIP_PERF=false

:parse_args
if "%~1"=="" goto end_parse_args
if /i "%~1"=="--quick" (
    set QUICK_MODE=true
    shift
    goto parse_args
)
if /i "%~1"=="--skip-build" (
    set SKIP_BUILD=true
    shift
    goto parse_args
)
if /i "%~1"=="--skip-perf" (
    set SKIP_PERF=true
    shift
    goto parse_args
)
if /i "%~1"=="--help" (
    echo Usage: %~nx0 [options]
    echo.
    echo Options:
    echo   --quick       Fast verification (skip race detection and integration tests^)
    echo   --skip-build  Skip the Wails build step
    echo   --skip-perf   Skip performance benchmarks
    echo   --help        Show this help message
    exit /b 0
)
echo ERROR: Unknown option: %~1
echo Use --help for usage information
exit /b 1

:end_parse_args

REM Color output helper (Windows 10+ with ANSI support)
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM Check prerequisites
echo.
echo %BLUE%==^>%NC% %GREEN%Checking prerequisites...%NC%

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Go is not installed. Please install Go 1.21 or later.
    exit /b 1
)

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Node.js is not installed. Please install Node.js 18 or later.
    exit /b 1
)

where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% npm is not installed. Please install npm.
    exit /b 1
)

if "%SKIP_BUILD%"=="false" (
    where wails >nul 2>nul
    if !errorlevel! neq 0 (
        echo %RED%ERROR:%NC% Wails is not installed. Please run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
        exit /b 1
    )
)

where staticcheck >nul 2>nul
if %errorlevel% neq 0 (
    echo %YELLOW%WARNING:%NC% staticcheck is not installed. Linting will be skipped.
    echo %YELLOW%WARNING:%NC% Install with: go install honnef.co/go/tools/cmd/staticcheck@latest
)

echo Prerequisites check passed.

REM Download Go dependencies
echo.
echo %BLUE%==^>%NC% %GREEN%Downloading Go dependencies...%NC%
go mod download
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Failed to download Go dependencies
    exit /b 1
)

REM Run Go formatting check
echo.
echo %BLUE%==^>%NC% %GREEN%Checking Go formatting...%NC%
for /f "delims=" %%i in ('gofmt -s -l . 2^>nul ^| findstr /v "vendor\\ frontend\\ build\\" ^| findstr "\.go$"') do (
    echo %RED%ERROR:%NC% Code is not formatted. Run 'go fmt ./...' to fix:
    gofmt -s -l . | findstr /v "vendor\\ frontend\\ build\\" | findstr "\.go$"
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% Go formatting check passed.

REM Run go vet
echo.
echo %BLUE%==^>%NC% %GREEN%Running go vet...%NC%
go vet ./...
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% go vet failed
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% go vet passed.

REM Run staticcheck
where staticcheck >nul 2>nul
if %errorlevel% equ 0 (
    echo.
    echo %BLUE%==^>%NC% %GREEN%Running staticcheck...%NC%
    staticcheck ./...
    if !errorlevel! neq 0 (
        echo %RED%ERROR:%NC% staticcheck failed
        exit /b 1
    )
    echo %GREEN%SUCCESS:%NC% staticcheck passed.
)

REM Run backend unit tests
echo.
echo %BLUE%==^>%NC% %GREEN%Running backend unit tests...%NC%
if "%QUICK_MODE%"=="true" (
    go test ./internal/... -v
) else (
    go test ./internal/... -v -race -coverprofile=coverage.txt -covermode=atomic
)
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Backend unit tests failed
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% Backend unit tests passed.

REM Run contract tests
echo.
echo %BLUE%==^>%NC% %GREEN%Running contract tests...%NC%
go test ./tests/contract/... -v
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Contract tests failed
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% Contract tests passed.

REM Run integration tests (unless in quick mode)
if "%QUICK_MODE%"=="false" (
    echo.
    echo %BLUE%==^>%NC% %GREEN%Running integration tests...%NC%
    go test ./tests/integration/... -v -timeout=5m
    if !errorlevel! neq 0 (
        echo %RED%ERROR:%NC% Integration tests failed
        exit /b 1
    )
    echo %GREEN%SUCCESS:%NC% Integration tests passed.
) else (
    echo %YELLOW%WARNING:%NC% Skipping integration tests (quick mode^).
)

REM Run performance benchmarks (optional)
if "%SKIP_PERF%"=="false" (
    if "%QUICK_MODE%"=="false" (
        echo.
        echo %BLUE%==^>%NC% %GREEN%Running performance benchmarks...%NC%
        go test ./tests/performance/... -v -timeout=10m
        if !errorlevel! neq 0 (
            echo %RED%ERROR:%NC% Performance benchmarks failed
            exit /b 1
        )
        echo %GREEN%SUCCESS:%NC% Performance benchmarks passed.
    ) else (
        echo %YELLOW%WARNING:%NC% Skipping performance benchmarks (quick mode^).
    )
) else (
    echo %YELLOW%WARNING:%NC% Skipping performance benchmarks (--skip-perf flag^).
)

REM Frontend checks
echo.
echo %BLUE%==^>%NC% %GREEN%Installing frontend dependencies...%NC%
cd frontend
call npm install
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Failed to install frontend dependencies
    cd ..
    exit /b 1
)

echo.
echo %BLUE%==^>%NC% %GREEN%Running TypeScript type check...%NC%
call npm run check
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% TypeScript type check failed
    cd ..
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% TypeScript type check passed.

echo.
echo %BLUE%==^>%NC% %GREEN%Running frontend tests...%NC%
call npm test
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Frontend tests failed
    cd ..
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% Frontend tests passed.

echo.
echo %BLUE%==^>%NC% %GREEN%Building frontend...%NC%
call npm run build
if %errorlevel% neq 0 (
    echo %RED%ERROR:%NC% Frontend build failed
    cd ..
    exit /b 1
)
echo %GREEN%SUCCESS:%NC% Frontend build passed.

cd ..

REM Build with Wails (optional)
if "%SKIP_BUILD%"=="false" (
    echo.
    echo %BLUE%==^>%NC% %GREEN%Building application with Wails...%NC%
    wails build -clean
    if !errorlevel! neq 0 (
        echo %RED%ERROR:%NC% Wails build failed
        exit /b 1
    )
    echo %GREEN%SUCCESS:%NC% Wails build passed.
) else (
    echo %YELLOW%WARNING:%NC% Skipping Wails build (--skip-build flag^).
)

REM Final summary
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  All verification checks passed!%NC%
echo %GREEN%========================================%NC%
echo.

if "%QUICK_MODE%"=="true" (
    echo %YELLOW%WARNING:%NC% Quick mode was used. Consider running full verification before pushing.
)

exit /b 0
