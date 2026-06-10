# ============================================================
# HRIS-Stery Preflight Checks (PowerShell)
# ============================================================
# Validates environment before running `make dev`
# Checks: Docker, Docker Compose, required binaries, disk space, ports

param(
    [switch]$Verbose = $false
)

$ErrorActionPreference = 'Continue'
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$ScriptName = Split-Path -Leaf $PSCommandPath

# Color codes (using Write-Host colors)
$ColorPass = 'Green'
$ColorFail = 'Red'
$ColorWarn = 'Yellow'
$ColorInfo = 'Cyan'

# Counters
$ChecksPassed = 0
$ChecksFailed = 0
$ChecksWarning = 0

# ──────────────────────────────────────────────────────────────
# Helper functions
# ──────────────────────────────────────────────────────────────

function Write-Check {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Message,

        [Parameter(Mandatory=$true)]
        [ValidateSet('Pass', 'Fail', 'Warning')]
        [string]$Status
    )

    switch ($Status) {
        'Pass' {
            Write-Host "✓ $Message" -ForegroundColor $ColorPass
            $script:ChecksPassed++
        }
        'Fail' {
            Write-Host "✗ $Message" -ForegroundColor $ColorFail
            $script:ChecksFailed++
        }
        'Warning' {
            Write-Host "⊘ $Message" -ForegroundColor $ColorWarn
            $script:ChecksWarning++
        }
    }
}

function Test-CommandExists {
    param([string]$Command)
    $null = Get-Command $Command -ErrorAction SilentlyContinue
    return $?
}

function Test-PortInUse {
    param(
        [int]$Port,
        [string]$Service
    )

    try {
        $tcpClient = [System.Net.Sockets.TcpClient]::new()
        $tcpClient.Connect('localhost', $Port)
        $tcpClient.Close()
        return $true  # Port is in use
    }
    catch {
        return $false  # Port is available
    }
}

function Test-NetConnection {
    param([string]$ComputerName)

    try {
        $result = Test-NetConnection -ComputerName $ComputerName -Port 443 -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
        return $result.TcpTestSucceeded
    }
    catch {
        return $false
    }
}

# ──────────────────────────────────────────────────────────────
# Check functions
# ──────────────────────────────────────────────────────────────

function Check-Docker {
    Write-Host "`nDocker & Docker Compose" -ForegroundColor $ColorInfo

    # Check Docker CLI
    if (-not (Test-CommandExists 'docker')) {
        Write-Check "Docker is not installed" -Status 'Fail'
        Write-Host "  Install from: https://www.docker.com/products/docker-desktop" -ForegroundColor $ColorWarn
        return $false
    }
    Write-Check "Docker CLI found" -Status 'Pass'

    # Check Docker daemon
    try {
        $null = docker info 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Check "Docker daemon is running" -Status 'Pass'
        }
        else {
            Write-Check "Docker daemon is not running" -Status 'Fail'
            Write-Host "  Start Docker Desktop or run: Set-Service docker -StartupType Automatic; Start-Service docker" -ForegroundColor $ColorWarn
            return $false
        }
    }
    catch {
        Write-Check "Docker daemon is not running" -Status 'Fail'
        return $false
    }

    # Check Docker Compose
    try {
        $version = docker compose version 2>&1 | Where-Object { $_ -match 'v\d+\.\d+\.\d+' } | Select-Object -First 1
        if ($version) {
            Write-Check "Docker Compose found" -Status 'Pass'
        }
        else {
            Write-Check "Docker Compose is not available" -Status 'Fail'
            return $false
        }
    }
    catch {
        Write-Check "Docker Compose is not available" -Status 'Fail'
        return $false
    }

    return $true
}

function Check-RequiredBinaries {
    Write-Host "`nRequired Binaries" -ForegroundColor $ColorInfo

    $required = @('go', 'buf', 'git')
    foreach ($cmd in $required) {
        if (Test-CommandExists $cmd) {
            Write-Check "$cmd found" -Status 'Pass'
        }
        else {
            Write-Check "$cmd is not installed" -Status 'Fail'
        }
    }
}

function Check-OptionalBinaries {
    Write-Host "`nOptional Binaries" -ForegroundColor $ColorInfo

    $optional = @(
        @{ cmd = 'golangci-lint'; note = 'for: make lint-go'; url = 'https://golangci-lint.run/usage/install/' },
        @{ cmd = 'migrate'; note = 'for: make migrate-*'; url = 'https://github.com/golang-migrate/migrate' },
        @{ cmd = 'node'; note = 'for: make dev-svelte'; url = 'https://nodejs.org/' },
        @{ cmd = 'php'; note = 'for: make dev-laravel'; url = 'https://www.php.net/downloads' },
        @{ cmd = 'composer'; note = 'for: make dev-laravel'; url = 'https://getcomposer.org/download/' }
    )

    foreach ($item in $optional) {
        if (Test-CommandExists $item.cmd) {
            Write-Check "$($item.cmd) found" -Status 'Pass'
        }
        else {
            Write-Check "$($item.cmd) not found (install for: $($item.note))" -Status 'Warning'
        }
    }
}

function Check-DiskSpace {
    Write-Host "`nDisk Space" -ForegroundColor $ColorInfo

    try {
        $drive = (Get-Item $ProjectRoot).PSDrive
        $freeGB = [math]::Round($drive.Free / 1GB, 2)

        if ($drive.Free -lt 10GB) {
            Write-Check "Insufficient disk space: ${freeGB}GB available (need ≥10GB)" -Status 'Fail'
            return $false
        }
        else {
            Write-Check "Sufficient disk space: ${freeGB}GB available" -Status 'Pass'
            return $true
        }
    }
    catch {
        Write-Check "Could not determine disk space" -Status 'Warning'
        return $true
    }
}

function Check-Ports {
    Write-Host "`nRequired Ports Available" -ForegroundColor $ColorInfo

    $ports = @(
        @{ port = 5432; service = 'PostgreSQL' },
        @{ port = 6379; service = 'Redis' },
        @{ port = 4222; service = 'NATS' },
        @{ port = 9000; service = 'MinIO API' },
        @{ port = 9001; service = 'MinIO Console' },
        @{ port = 8000; service = 'Laravel' },
        @{ port = 5173; service = 'SvelteKit (dev)' },
        @{ port = 4173; service = 'SvelteKit (preview)' }
    )

    foreach ($p in $ports) {
        if (Test-PortInUse -Port $p.port -Service $p.service) {
            Write-Check "Port $($p.port) ($($p.service)) is in use" -Status 'Fail'
        }
        else {
            Write-Check "Port $($p.port) ($($p.service)) available" -Status 'Pass'
        }
    }
}

function Check-Network {
    Write-Host "`nNetwork & Registries" -ForegroundColor $ColorInfo

    # Check GitHub
    if (Test-NetConnection -ComputerName 'github.com') {
        Write-Check "Can reach github.com" -Status 'Pass'
    }
    else {
        Write-Check "Cannot reach github.com (may fail downloading Go modules)" -Status 'Warning'
    }

    # Check Docker Hub
    if (Test-NetConnection -ComputerName 'docker.io') {
        Write-Check "Can reach docker.io" -Status 'Pass'
    }
    else {
        Write-Check "Cannot reach docker.io (may fail pulling images)" -Status 'Warning'
    }
}

function Check-DockerComposeFile {
    Write-Host "`nDocker Compose Configuration" -ForegroundColor $ColorInfo

    $composeFile = Join-Path $ProjectRoot 'deploy' 'docker' 'docker-compose.yml'

    if (-not (Test-Path $composeFile)) {
        Write-Check "docker-compose.yml not found at: $composeFile" -Status 'Fail'
        return $false
    }
    Write-Check "docker-compose.yml found" -Status 'Pass'

    # Validate docker-compose file
    try {
        $null = docker compose -f $composeFile config 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Check "docker-compose.yml is valid" -Status 'Pass'
            return $true
        }
        else {
            Write-Check "docker-compose.yml has errors" -Status 'Fail'
            return $false
        }
    }
    catch {
        Write-Check "docker-compose.yml has errors" -Status 'Fail'
        return $false
    }
}

# ──────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────

function Main {
    Write-Host "`n" + ("=" * 55) -ForegroundColor $ColorInfo
    Write-Host "  HRIS-Stery Preflight Checks" -ForegroundColor $ColorInfo
    Write-Host ("=" * 55) -ForegroundColor $ColorInfo

    Check-Docker
    Check-RequiredBinaries
    Check-OptionalBinaries
    Check-DiskSpace
    Check-Ports
    Check-Network
    Check-DockerComposeFile

    # Print summary
    Write-Host "`n" + ("=" * 55) -ForegroundColor $ColorInfo
    Write-Host "Checks Summary:"
    Write-Host "  Passed:   $ChecksPassed" -ForegroundColor $ColorPass
    Write-Host "  Warnings: $ChecksWarning" -ForegroundColor $ColorWarn
    Write-Host "  Failed:   $ChecksFailed" -ForegroundColor $ColorFail
    Write-Host ("=" * 55) -ForegroundColor $ColorInfo
    Write-Host ""

    if ($ChecksFailed -gt 0) {
        Write-Host "❌ Preflight checks FAILED" -ForegroundColor $ColorFail
        Write-Host "Please fix the above issues and run 'make dev' again."
        exit 1
    }

    if ($ChecksWarning -gt 0) {
        Write-Host "⚠️  Preflight checks PASSED with warnings" -ForegroundColor $ColorWarn
        Write-Host "Some optional features may not be available."
        Write-Host "Run 'make dev' to continue."
    }
    else {
        Write-Host "✅ All preflight checks PASSED" -ForegroundColor $ColorPass
        Write-Host "Ready to run: make dev"
    }

    exit 0
}

Main
