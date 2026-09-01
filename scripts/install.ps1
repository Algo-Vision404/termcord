# Build from source
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Push-Location $Root
try {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go is required: https://go.dev/dl/"
    }
    go build -o bin/termcord.exe ./cmd/termcord
    go build -o bin/termcord-cli.exe ./cmd/termcord-cli
    Write-Host "Installed to $Root\bin"
    Write-Host "Next: .\bin\termcord.exe init && .\bin\termcord.exe login"
} finally {
    Pop-Location
}
