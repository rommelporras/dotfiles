# Windows Git Setup
# Run from PowerShell or Windows Terminal on a fresh Windows machine.
# Configures Git for Windows to use the Windows OpenSSH agent (required for
# 1Password SSH agent integration).

$ErrorActionPreference = "Stop"

Write-Host "=== Windows Git Setup ===" -ForegroundColor Cyan

# Verify git is available
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "Error: git not found. Install Git for Windows first." -ForegroundColor Red
    exit 1
}

# Point Git at Windows OpenSSH so it uses the Windows SSH agent (1Password)
$sshPath = "C:/Windows/System32/OpenSSH/ssh.exe"
if (-not (Test-Path $sshPath)) {
    Write-Host "Error: Windows OpenSSH not found at $sshPath" -ForegroundColor Red
    Write-Host "Enable it via Settings > Apps > Optional Features > OpenSSH Client"
    exit 1
}

$current = git config --global --get core.sshCommand 2>$null
if ($current -eq $sshPath) {
    Write-Host "core.sshCommand already set to $sshPath - skipping" -ForegroundColor Yellow
} else {
    git config --global core.sshCommand $sshPath
    Write-Host "Set core.sshCommand = $sshPath" -ForegroundColor Green
}

# Verify
Write-Host ""
Write-Host "=== Current Windows global git config ===" -ForegroundColor Cyan
git config --global --list --show-origin

Write-Host ""
Write-Host "Done. Git for Windows will now use the Windows SSH agent (1Password)." -ForegroundColor Green
