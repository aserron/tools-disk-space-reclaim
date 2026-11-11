<#
.SYNOPSIS
    Builds and tests the space-reclaim CLI utility.
.DESCRIPTION
    Runs unit tests, builds the binary, and places space-reclaim.exe at the project root.
.PARAMETER Output
    Optional custom destination path for the binary.
#>
param(
    [string]$Output = "$PSScriptRoot\..\space-reclaim.exe"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Resolve-Path "$PSScriptRoot\.."

# Use a work directory on E:\ for go test binaries so Windows App Control policies on %TEMP% do not block test executables
$GoTmp = Join-Path $ProjectRoot ".gotmp"
New-Item -ItemType Directory -Force -Path $GoTmp | Out-Null
$env:GOTMPDIR = $GoTmp

try {
    Write-Host "==> Running unit tests..." -ForegroundColor Cyan
    Set-Location $ProjectRoot
    go test ./... -v

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Tests failed!"
        exit 1
    }

    Write-Host "`n==> Building space-reclaim..." -ForegroundColor Cyan
    go build -ldflags="-s -w" -o $Output ./cmd/space-reclaim

    if (Test-Path $Output) {
        $size = (Get-Item $Output).Length / 1MB
        Write-Host "`n[SUCCESS] Built binary: $Output ($([math]::Round($size, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Error "Build failed!"
        exit 1
    }
} finally {
    Remove-Item -Path $GoTmp -Recurse -Force -ErrorAction SilentlyContinue
}
