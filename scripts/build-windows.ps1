$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$dist = Join-Path $root "dist"
$exe = Join-Path $dist "yaml-struct-editor.exe"

New-Item -ItemType Directory -Force -Path $dist | Out-Null

Push-Location (Join-Path $root "frontend")
npm run build
Pop-Location

Push-Location $root
go build -trimpath -ldflags="-w -s -H windowsgui" -o $exe ./cmd/yaml-struct-editor
Pop-Location

Write-Host "Built: $exe"
