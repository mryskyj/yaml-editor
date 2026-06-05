$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$frontendIndex = Join-Path $root "frontend\dist\index.html"
$vendorModules = Join-Path $root "vendor\modules.txt"
$dist = Join-Path $root "dist"
$exe = Join-Path $dist "yaml-struct-editor.exe"

if (-not (Test-Path $frontendIndex)) {
    throw "frontend\dist is missing. Run npm run build on an online development machine and include frontend\dist before running this script."
}

if (-not (Test-Path $vendorModules)) {
    throw "vendor\modules.txt is missing. Run go mod vendor on an online development machine and include vendor before running this script."
}

New-Item -ItemType Directory -Force -Path $dist | Out-Null

Push-Location $root
go build -mod=vendor -tags production -trimpath -buildvcs=false -ldflags="-w -s -H windowsgui" -o $exe ./cmd/yaml-struct-editor
Pop-Location

Write-Host "Built: $exe"
