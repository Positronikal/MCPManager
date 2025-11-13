#!/usr/bin/env pwsh
<#
.SYNOPSIS
Update vendored Spec Kit scripts to a new version

.DESCRIPTION
This script updates the vendored Spec Kit scripts in .specify/scripts/powershell/
from the local spec-kit repository. It creates a backup, copies the new version,
and updates VERSION.txt.

.PARAMETER Version
The version string to record (e.g., "v1.1.0", "v2.0.0")

.PARAMETER SpecKitPath
Path to the spec-kit repository (default: D:\dev\ARTIFICIAL_INTELLIGENCE\spec-kit)

.PARAMETER Force
Skip confirmation prompt

.EXAMPLE
.\update-spec-kit.ps1 -Version v1.1.0

.EXAMPLE
.\update-spec-kit.ps1 -Version v1.2.0 -SpecKitPath "C:\repos\spec-kit" -Force

.NOTES
This script should only be run by project maintainers after testing spec-kit updates.
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    
    [string]$SpecKitPath = "D:\dev\ARTIFICIAL_INTELLIGENCE\spec-kit",
    
    [switch]$Force
)

$ErrorActionPreference = 'Stop'

# Color output functions
function Write-Info { param([string]$Msg) Write-Host "ℹ️  $Msg" -ForegroundColor Cyan }
function Write-Success { param([string]$Msg) Write-Host "✅ $Msg" -ForegroundColor Green }
function Write-Warning { param([string]$Msg) Write-Host "⚠️  $Msg" -ForegroundColor Yellow }
function Write-Err { param([string]$Msg) Write-Host "❌ $Msg" -ForegroundColor Red }

# Validate we're in the right place
$projectRoot = Split-Path -Parent $PSScriptRoot | Split-Path -Parent | Split-Path -Parent
if (-not (Test-Path (Join-Path $projectRoot ".specify\VERSION.txt"))) {
    Write-Err "This script must be run from the .specify/scripts/powershell/ directory"
    exit 1
}

# Validate spec-kit path
if (-not (Test-Path $SpecKitPath)) {
    Write-Err "Spec Kit path not found: $SpecKitPath"
    Write-Info "Specify the correct path with -SpecKitPath parameter"
    exit 1
}

$sourceScripts = Join-Path $SpecKitPath "scripts\powershell"
if (-not (Test-Path $sourceScripts)) {
    Write-Err "Spec Kit scripts directory not found: $sourceScripts"
    exit 1
}

# Read current version
$versionFile = Join-Path $projectRoot ".specify\VERSION.txt"
$currentVersion = (Get-Content $versionFile -First 1).Replace("spec-kit ", "").Trim()

Write-Info "=== Spec Kit Update Utility ==="
Write-Host ""
Write-Info "Current version: $currentVersion"
Write-Info "Target version:  $Version"
Write-Info "Source path:     $SpecKitPath"
Write-Host ""

# Confirmation prompt (unless -Force)
if (-not $Force) {
    Write-Warning "This will:"
    Write-Host "  1. Create backup of current scripts"
    Write-Host "  2. Copy new scripts from spec-kit"
    Write-Host "  3. Update VERSION.txt"
    Write-Host ""
    Write-Warning "You MUST test thoroughly before committing!"
    Write-Host ""
    
    $response = Read-Host "Continue? (yes/no)"
    if ($response -ne "yes") {
        Write-Info "Update cancelled"
        exit 0
    }
}

# Create backup
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$backupDir = Join-Path $projectRoot ".specify\scripts.backup-$timestamp"

Write-Info "Creating backup at: $backupDir"
Copy-Item (Join-Path $projectRoot ".specify\scripts") $backupDir -Recurse -Force
Write-Success "Backup created"

# Copy new scripts
Write-Info "Copying scripts from spec-kit..."

$scriptFiles = @(
    "check-prerequisites.ps1",
    "common.ps1",
    "create-new-feature.ps1",
    "setup-plan.ps1",
    "update-agent-context.ps1"
)

$destScripts = Join-Path $projectRoot ".specify\scripts\powershell"
$copyCount = 0

foreach ($script in $scriptFiles) {
    $source = Join-Path $sourceScripts $script
    $dest = Join-Path $destScripts $script
    
    if (Test-Path $source) {
        Copy-Item $source $dest -Force
        Write-Host "  ✓ $script" -ForegroundColor Gray
        $copyCount++
    } else {
        Write-Warning "Script not found in spec-kit: $script (skipped)"
    }
}

Write-Success "Copied $copyCount scripts"

# Update VERSION.txt
Write-Info "Updating VERSION.txt..."

$versionContent = @"
spec-kit $Version
Vendored: $(Get-Date -Format 'yyyy-MM-dd')
Source: https://github.com/github/spec-kit
Maintainer: Hoyt

DO NOT UPDATE MANUALLY - See project documentation for update process

Changelog:
- $(Get-Date -Format 'yyyy-MM-dd'): Updated to spec-kit $Version
"@

# Preserve old changelog entries (keep last 3 updates)
$oldContent = Get-Content $versionFile -Raw
if ($oldContent -match '(?s)Changelog:\r?\n(.+)') {
    $oldChangelog = $matches[1]
    $oldEntries = ($oldChangelog -split '(?=^- \d{4}-\d{2}-\d{2})' | Where-Object { $_ -match '^\s*-' } | Select-Object -Skip 1 -First 2) -join ""
    if ($oldEntries) {
        $versionContent += $oldEntries
    }
}

Set-Content -Path $versionFile -Value $versionContent -Encoding UTF8
Write-Success "VERSION.txt updated"

# Summary
Write-Host ""
Write-Success "=== Update Complete ==="
Write-Host ""
Write-Warning "NEXT STEPS (IMPORTANT):"
Write-Host "  1. Review changes:  git diff .specify/" -ForegroundColor White
Write-Host "  2. Test thoroughly: Run your project's test suite" -ForegroundColor White
Write-Host "  3. If broken:       git checkout .specify/ to restore" -ForegroundColor White
Write-Host "  4. If working:      git add .specify/ && git commit" -ForegroundColor White
Write-Host ""
Write-Info "Backup available at: $backupDir"
Write-Host ""

exit 0
