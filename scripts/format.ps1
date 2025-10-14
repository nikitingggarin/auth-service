#!/bin/pwsh

Write-Host "Formatting Go code with gofumpt..." -ForegroundColor Cyan

# орматируем весь проект
gofumpt -w .

# рганизуем импорты
goimports -w .

Write-Host "Code formatted successfully!" -ForegroundColor Green
