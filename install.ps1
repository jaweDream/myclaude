[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'Continue'
$tempFile = $null

function Get-Architecture {
    $arch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
    switch ($arch.ToLower()) {
        'amd64' { 'amd64' }
        'x86' { throw 'Unsupported architecture: x86 (64-bit Windows is required).' }
        'arm64' { 'arm64' }
        'aarch64' { 'arm64' }
        default { throw "Unsupported architecture: $arch" }
    }
}

function Write-Step {
    param([string] $Status, [int] $Percent)
    Write-Progress -Activity 'Installing codex-wrapper' -Status $Status -PercentComplete $Percent
    Write-Host $Status
}

try {
    Write-Step 'Detecting CPU architecture...' 5
    $arch = Get-Architecture
    $url = "https://github.com/cexll/myclaude/releases/latest/download/codex-wrapper-windows-$arch.exe"

    $tempFile = Join-Path ([IO.Path]::GetTempPath()) "codex-wrapper-$arch.exe"
    $homeBin = Join-Path $HOME 'bin'
    $destination = Join-Path $homeBin 'codex-wrapper.exe'

    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    Write-Step "Downloading codex-wrapper from $url ..." 25
    Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing -ErrorAction Stop

    Write-Step "Installing to $destination ..." 65
    New-Item -ItemType Directory -Path $homeBin -Force | Out-Null
    Move-Item -LiteralPath $tempFile -Destination $destination -Force

    Write-Step 'Verifying installation...' 90
    & $destination --version | Out-Null

    Write-Step 'codex-wrapper installed successfully.' 100
    Write-Progress -Activity 'Installing codex-wrapper' -Completed -Status 'Done'
    Write-Host "Installed to $destination"

    $normalizedBin = ($homeBin.TrimEnd('\') -replace '/','\').ToLower()
    $pathEntries = ($env:PATH -split ';') | Where-Object { $_ } | ForEach-Object { ($_ -replace '/','\').TrimEnd('\').ToLower() }
    if (-not ($pathEntries -contains $normalizedBin)) {
        Write-Warning "$homeBin is not in your PATH."
        Write-Host 'Add it permanently with:'
        Write-Host "  [Environment]::SetEnvironmentVariable('PATH', '$homeBin;' + [Environment]::GetEnvironmentVariable('PATH','User'), 'User')"
        Write-Host 'Then restart your shell to pick up the updated PATH.'
    }
} catch {
    Write-Progress -Activity 'Installing codex-wrapper' -Completed -Status 'Failed'
    Write-Error "Installation failed: $($_.Exception.Message)"
    exit 1
} finally {
    if ($tempFile -and (Test-Path $tempFile)) {
        Remove-Item -LiteralPath $tempFile -Force -ErrorAction SilentlyContinue
    }
}
