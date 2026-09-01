# Publish termcord to GitHub and create v0.1.0 release.
# Run from repo root: .\scripts\publish-github.ps1

$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)

Write-Host "==> Checking git status..."
git status --porcelain
if ($LASTEXITCODE -ne 0) { throw "git status failed" }

Write-Host "==> Ensuring no secrets staged..."
$staged = git diff --cached --name-only 2>$null
foreach ($f in @("config.toml", "token.txt", ".env")) {
    if ($staged -match [regex]::Escape($f)) {
        throw "Refusing to publish: $f is staged"
    }
}

Write-Host "==> Committing changes..."
$dirty = git status --porcelain
if ($dirty) {
    git add -A
    git commit -m @"
Release v0.1.0: design system, gateway fixes, and security hardening.

Includes stream-layout TUI, gateway reliability, plugin/cache security,
CI with govulncheck, SECURITY.md, and semver release docs.
"@
}

Write-Host "==> Renaming branch to main..."
git branch -M main

Write-Host "==> Checking gh auth..."
gh auth status
if ($LASTEXITCODE -ne 0) {
    Write-Host "Run: gh auth login"
    exit 1
}

Write-Host "==> Creating public GitHub repo (if needed)..."
$remote = git remote get-url origin 2>$null
if (-not $remote) {
    gh repo create termcord --public --source=. --remote=origin --description "Terminal-native Discord client — direct gateway, local cache, keyboard-first TUI"
    if ($LASTEXITCODE -ne 0) {
        $user = gh api user -q .login
        gh repo create "$user/termcord" --public --source=. --remote=origin --description "Terminal-native Discord client — direct gateway, local cache, keyboard-first TUI"
    }
}

Write-Host "==> Pushing..."
git push -u origin main

Write-Host "==> Building release binaries..."
New-Item -ItemType Directory -Force -Path dist | Out-Null
$version = "0.1.0"
$ldflags = "-s -w -X github.com/termcord/termcord/internal/version.Version=$version"
go build -ldflags $ldflags -o "dist/termcord-windows-amd64.exe" ./cmd/termcord
go build -ldflags $ldflags -o "dist/termcord-cli-windows-amd64.exe" ./cmd/termcord-cli

Write-Host "==> Creating GitHub release v0.1.0..."
gh release view v0.1.0 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "Release v0.1.0 already exists — refreshing notes and uploading assets."
    gh release edit v0.1.0 --notes-file docs/RELEASE_NOTES_v0.1.0.md
    gh release upload v0.1.0 dist/* --clobber
} else {
    gh release create v0.1.0 `
        --title "v0.1.0 — First public release" `
        --notes-file docs/RELEASE_NOTES_v0.1.0.md `
        dist/*
}

Write-Host ""
Write-Host "Done."
gh repo view --web 2>$null
Write-Host "Release: $(gh release view v0.1.0 --json url -q .url)"
