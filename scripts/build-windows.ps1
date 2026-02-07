# Build VoiceCode for Windows.
# Usage: .\scripts\build-windows.ps1
$ErrorActionPreference = "Stop"

$AppName = "VoiceCode"
$BuildDir = "build"

Write-Host "Building $AppName for Windows..."

if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
}

# Build with -H windowsgui to suppress console window
go build -ldflags="-s -w -H windowsgui" -o "$BuildDir\voicecode.exe" .\cmd\voicecode

Write-Host "Built: $BuildDir\voicecode.exe"
Get-Item "$BuildDir\voicecode.exe" | Select-Object Name, Length
