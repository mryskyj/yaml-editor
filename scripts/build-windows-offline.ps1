$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$frontendIndex = Join-Path $root "frontend\dist\index.html"
$dist = Join-Path $root "dist"
$exe = Join-Path $dist "yaml-struct-editor.exe"

if (-not (Test-Path $frontendIndex)) {
    throw "frontend\dist is missing. Run npm run build on an online development machine and include frontend\dist before running this script."
}

New-Item -ItemType Directory -Force -Path $dist | Out-Null

Push-Location $root
go build -tags production -trimpath -buildvcs=false -ldflags="-w -s -H windowsgui" -o $exe ./cmd/yaml-struct-editor
Pop-Location

Write-Host "Built: $exe"
