# Coverage Report Generator for FreeStealer
# Run this script to generate comprehensive coverage reports

Write-Host "=== FreeStealer Coverage Report ===" -ForegroundColor Cyan
Write-Host ""

# Clean old coverage files
Write-Host "Cleaning old coverage files..." -ForegroundColor Yellow
Remove-Item -Path "coverage.out" -ErrorAction SilentlyContinue
Remove-Item -Path "coverage.html" -ErrorAction SilentlyContinue

# Run tests with coverage
Write-Host "Running tests with coverage..." -ForegroundColor Yellow
go test ./... -coverprofile=coverage.out -covermode=atomic

if ($LASTEXITCODE -ne 0) {
    Write-Host "Tests failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Coverage Summary ===" -ForegroundColor Green

# Display total coverage
$coverageOutput = go tool cover -func coverage.out
$totalCoverage = $coverageOutput | Select-Object -Last 1
Write-Host $totalCoverage -ForegroundColor White

Write-Host ""
Write-Host "=== Per-Package Coverage ===" -ForegroundColor Green

# Get coverage per package
$packages = go list ./...
foreach ($pkg in $packages) {
    $pkgName = $pkg -replace "freestealer/", ""
    $coverage = go test -cover $pkg 2>&1 | Select-String -Pattern "coverage:"
    if ($coverage) {
        Write-Host "$pkgName : $coverage" -ForegroundColor Cyan
    }
}

# Generate HTML report
Write-Host ""
Write-Host "Generating HTML coverage report..." -ForegroundColor Yellow
go tool cover -html coverage.out -o coverage.html

Write-Host ""
Write-Host "=== Coverage Report Generated ===" -ForegroundColor Green
Write-Host "HTML Report: coverage.html" -ForegroundColor Cyan
Write-Host "Profile Data: coverage.out" -ForegroundColor Cyan
Write-Host ""
Write-Host "To view HTML report, run: Start-Process coverage.html" -ForegroundColor Yellow
