# Validates Makefile targets on Windows (PowerShell).
# Usage: .\scripts\validate-makefile.ps1
$ErrorActionPreference = "Continue"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

$LogDir = Join-Path $Root ".make-validate"
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
$Results = Join-Path $LogDir "results.tsv"
"target`texit`tstatus" | Set-Content $Results

function Run-Target {
    param(
        [string]$Target,
        [switch]$AllowFail
    )
    $log = Join-Path $LogDir "$Target.log"
    Write-Host "======== make $Target ========"
    cmd /c "make $Target > `"$log`" 2>&1"
    $code = $LASTEXITCODE
    $status = if ($code -eq 0) { "PASS" } elseif ($AllowFail) { "EXPECTED_FAIL" } else { "FAIL" }
    "$Target`t$code`t$status" | Add-Content $Results
    Get-Content $log -Tail 5 | ForEach-Object { Write-Host "  $_" }
    Write-Host ""
}

Run-Target help
Run-Target check-tools
Run-Target go-services
Run-Target proto-deps
Run-Target proto-lint
Run-Target proto-gen
Run-Target proto-breaking -AllowFail
Run-Target tidy
Run-Target vet
Run-Target test-go
Run-Target build-go
Run-Target lint-go -AllowFail
Run-Target test-laravel -AllowFail
Run-Target test-svelte -AllowFail
Run-Target lint-laravel -AllowFail
Run-Target lint-svelte -AllowFail
Run-Target build-svelte -AllowFail
Run-Target clean
Run-Target test -AllowFail
Run-Target lint -AllowFail
Run-Target build -AllowFail
Run-Target dev-ps -AllowFail
Run-Target migrate-status -AllowFail

Write-Host "Results: $Results"
Get-Content $Results | Format-Table
