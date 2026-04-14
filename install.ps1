# pvm installer for Windows
# Usage:
#   irm https://raw.githubusercontent.com/Higangssh/pvm/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo    = 'Higangssh/pvm'
$Binary  = 'pvm.exe'
$InstallDir = Join-Path $env:LOCALAPPDATA 'pvm'

Write-Host "==> Installing pvm..." -ForegroundColor Cyan

# Resolve latest release
$release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$tag = $release.tag_name
$asset = $release.assets | Where-Object { $_.name -eq $Binary } | Select-Object -First 1
if (-not $asset) {
    throw "Asset '$Binary' not found in release $tag"
}

Write-Host "    Version: $tag" -ForegroundColor Gray
Write-Host "    URL:     $($asset.browser_download_url)" -ForegroundColor Gray

# Download
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$target = Join-Path $InstallDir $Binary
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $target -UseBasicParsing
Write-Host "==> Downloaded to $target" -ForegroundColor Green

# Add to PATH (user scope) if missing
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not ($userPath -split ';' | Where-Object { $_ -eq $InstallDir })) {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$InstallDir", 'User')
    Write-Host "==> Added $InstallDir to user PATH" -ForegroundColor Green
    Write-Host "    Restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
} else {
    Write-Host "==> $InstallDir already in PATH" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Installed successfully!" -ForegroundColor Green
Write-Host "Try: pvm --help" -ForegroundColor Cyan
