# pvm uninstaller for Windows
# Usage:
#   irm https://raw.githubusercontent.com/Higangssh/pvm/main/uninstall.ps1 | iex

$ErrorActionPreference = 'Stop'

$InstallDir = Join-Path $env:LOCALAPPDATA 'pvm'
$ConfigDir  = Join-Path $env:APPDATA     'pvm'

Write-Host "==> Uninstalling pvm..." -ForegroundColor Cyan

# Remove binary directory
if (Test-Path $InstallDir) {
    Remove-Item -Recurse -Force $InstallDir
    Write-Host "    Removed $InstallDir" -ForegroundColor Green
} else {
    Write-Host "    $InstallDir not found (skipped)" -ForegroundColor Gray
}

# Remove from user PATH
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$newPath = ($userPath -split ';' | Where-Object { $_ -and $_ -ne $InstallDir }) -join ';'
if ($newPath -ne $userPath) {
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "    Removed $InstallDir from user PATH" -ForegroundColor Green
}

# Ask about config
if (Test-Path $ConfigDir) {
    $ans = Read-Host "Remove config at $ConfigDir too? (y/N)"
    if ($ans -eq 'y' -or $ans -eq 'Y') {
        Remove-Item -Recurse -Force $ConfigDir
        Write-Host "    Removed $ConfigDir" -ForegroundColor Green
    } else {
        Write-Host "    Kept $ConfigDir" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Uninstalled. Restart your terminal for PATH changes to apply." -ForegroundColor Cyan
