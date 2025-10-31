@echo off
chcp 65001 >nul
if "%1"=="" goto help

if "%1"=="run" (
    echo Starting application...
    go run cmd/server/main.go cmd/server/goroutine.go
    goto end
)

if "%1"=="build" (
    echo Building binary...
    if not exist bin mkdir bin
    go build -o bin/auth-service cmd/server/main.go cmd/server/goroutine.go
    echo Binary created: bin/auth-service
    goto end
)

if "%1"=="test" (
    echo Running tests...
    go test ./...
    goto end
)

if "%1"=="lint" (
    echo Running linter...
    golangci-lint run
    goto end
)

if "%1"=="format" (
    echo Formatting code...
    powershell -File scripts\format.ps1
    echo Code formatted
    goto end
)

if "%1"=="migrate-up" (
    echo Applying DB migrations...
    docker-compose exec postgres psql -U postgres -d auth_service -f /docker-entrypoint-initdb.d/001_create_users_table.up.sql
    echo Migrations applied
    goto end
)

if "%1"=="migrate-down" (
    echo Rolling back DB migrations...
    docker-compose exec postgres psql -U postgres -d auth_service -f /docker-entrypoint-initdb.d/001_create_users_table.down.sql
    echo Migrations rolled back
    goto end
)

if "%1"=="clean" (
    echo Cleaning binaries...
    if exist bin rmdir /s /q bin
    echo Bin folder cleaned
    goto end
)

:help
echo Available commands:
echo   make run         - Start application
echo   make build       - Build binary
echo   make test        - Run tests
echo   make lint        - Run linter
echo   make format      - Format code
echo   make migrate-up  - Apply DB migrations
echo   make migrate-down - Rollback DB migrations
echo   make clean       - Clean binaries

:end