param(
  [string]$ConfigPath = (Join-Path $PSScriptRoot 'config.psd1'),
  [switch]$SkipBackend,
  [switch]$SkipAdmin,
  [switch]$SkipMerchantH5,
  [switch]$ForceNginxReload,
  [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path $PSScriptRoot -Parent
$ArtifactRoot = Join-Path $PSScriptRoot '.artifacts'

function Write-Step {
  param([string]$Message)
  Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Fail {
  param([string]$Message)
  throw $Message
}

function Require-Command {
  param([string]$Name)
  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    Fail "Command not found: $Name"
  }
}

function Resolve-RepoPath {
  param([string]$RelativePath)
  if ([System.IO.Path]::IsPathRooted($RelativePath)) {
    return [System.IO.Path]::GetFullPath($RelativePath)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $RelativePath))
}

function Quote-Sh {
  param([string]$Value)
  if ($Value.Contains("'")) {
    Fail "Shell values containing single quotes are not supported: $Value"
  }
  return "'$Value'"
}

function Get-RemoteDirectoryName {
  param([string]$Path)

  $normalized = ($Path -replace '\\', '/').TrimEnd('/')
  if ([string]::IsNullOrWhiteSpace($normalized)) {
    return '/'
  }

  $lastSlashIndex = $normalized.LastIndexOf('/')
  if ($lastSlashIndex -lt 0) {
    return '.'
  }
  if ($lastSlashIndex -eq 0) {
    return '/'
  }

  return $normalized.Substring(0, $lastSlashIndex)
}

function Invoke-Native {
  param(
    [string]$FilePath,
    [string[]]$Arguments,
    [string]$WorkingDirectory
  )

  $display = if ($Arguments) { "$FilePath $($Arguments -join ' ')" } else { $FilePath }
  if ($DryRun) {
    Write-Host "[dry-run] $display"
    return
  }

  Push-Location $WorkingDirectory
  try {
    & $FilePath @Arguments | Out-Host
    if ($LASTEXITCODE -ne 0) {
      Fail "Command failed with exit code ${LASTEXITCODE}: $display"
    }
  }
  finally {
    Pop-Location
  }
}

function Invoke-WithEnvironmentOverrides {
  param(
    [hashtable]$Variables,
    [scriptblock]$Action
  )

  if (-not $Variables -or $Variables.Count -eq 0) {
    & $Action
    return
  }

  $previousValues = @{}
  foreach ($entry in $Variables.GetEnumerator()) {
    $name = [string]$entry.Key
    $existing = Get-Item -Path "Env:$name" -ErrorAction SilentlyContinue
    $previousValues[$name] = if ($existing) { $existing.Value } else { $null }
    Set-Item -Path "Env:$name" -Value ([string]$entry.Value)
  }

  try {
    & $Action
  }
  finally {
    foreach ($name in $Variables.Keys) {
      if ($null -eq $previousValues[$name]) {
        Remove-Item -Path "Env:$name" -ErrorAction SilentlyContinue
      }
      else {
        Set-Item -Path "Env:$name" -Value $previousValues[$name]
      }
    }
  }
}

function New-CleanDirectory {
  param([string]$Path)
  if (Test-Path $Path) {
    Remove-Item -Path $Path -Recurse -Force
  }
  New-Item -ItemType Directory -Path $Path | Out-Null
}

function Copy-DirectoryContents {
  param(
    [string]$SourceDir,
    [string]$DestinationDir,
    [string[]]$ExcludePaths = @()
  )

  $normalizedExcludes = @($ExcludePaths | Where-Object { $_ } | ForEach-Object { [System.IO.Path]::GetFullPath($_) })
  Get-ChildItem -LiteralPath $SourceDir -Force | ForEach-Object {
    $sourcePath = [System.IO.Path]::GetFullPath($_.FullName)
    if ($normalizedExcludes -contains $sourcePath) {
      return
    }
    Copy-Item -Path $sourcePath -Destination (Join-Path $DestinationDir $_.Name) -Recurse -Force
  }
}

function Get-SshTarget {
  param([hashtable]$SshConfig)
  return "$($SshConfig.User)@$($SshConfig.Host)"
}

function Get-SshArgs {
  param([hashtable]$SshConfig)
  $sshArgs = @('-p', [string]$SshConfig.Port)
  if ($SshConfig.KeyPath) {
    $sshArgs += @('-i', $SshConfig.KeyPath)
  }
  return $sshArgs
}

function Invoke-SshCommand {
  param(
    [hashtable]$SshConfig,
    [string]$Command
  )

  $sshArgs = Get-SshArgs -SshConfig $SshConfig
  $sshTarget = Get-SshTarget -SshConfig $SshConfig

  $display = "ssh $($sshArgs -join ' ') $sshTarget bash -s"
  if ($DryRun) {
    Write-Host "[dry-run] $display"
    Write-Host "[dry-run][stdin] $Command"
    return
  }

  $normalizedCommand = (($Command -replace "`r`n", "`n") -replace "`r", "`n")
  $processInfo = New-Object System.Diagnostics.ProcessStartInfo
  $processInfo.FileName = 'ssh'
  foreach ($arg in ($sshArgs + @($sshTarget, 'bash -s'))) {
    [void]$processInfo.ArgumentList.Add($arg)
  }
  $processInfo.WorkingDirectory = $RepoRoot
  $processInfo.UseShellExecute = $false
  $processInfo.RedirectStandardInput = $true
  $processInfo.RedirectStandardOutput = $true
  $processInfo.RedirectStandardError = $true

  $process = New-Object System.Diagnostics.Process
  $process.StartInfo = $processInfo
  try {
    [void]$process.Start()
    $process.StandardInput.NewLine = "`n"
    $process.StandardInput.WriteLine($normalizedCommand)
    $process.StandardInput.Close()

    $stdout = $process.StandardOutput.ReadToEnd()
    $stderr = $process.StandardError.ReadToEnd()
    $process.WaitForExit()

    if ($stdout) {
      Write-Host $stdout.TrimEnd()
    }
    if ($stderr) {
      Write-Host $stderr.TrimEnd()
    }

    if ($process.ExitCode -ne 0) {
      Fail "Command failed with exit code $($process.ExitCode): $display`n$normalizedCommand"
    }
  }
  finally {
    if ($process) {
      $process.Dispose()
    }
  }
}

function Copy-ToRemote {
  param(
    [hashtable]$SshConfig,
    [string]$LocalPath,
    [string]$RemotePath
  )

  $scpArgs = @('-P', [string]$SshConfig.Port)
  if ($SshConfig.KeyPath) {
    $scpArgs += @('-i', $SshConfig.KeyPath)
  }
  $sshTarget = Get-SshTarget -SshConfig $SshConfig
  $scpArgs += @($LocalPath, "${sshTarget}:$RemotePath")
  Invoke-Native -FilePath 'scp' -Arguments $scpArgs -WorkingDirectory $RepoRoot
}

function New-TarArchive {
  param(
    [string]$ArchivePath,
    [string[]]$Items,
    [string]$WorkingDirectory
  )

  if (Test-Path $ArchivePath) {
    Remove-Item -Path $ArchivePath -Force
  }
  $tarArgs = @('-czf', $ArchivePath) + $Items
  Invoke-Native -FilePath 'tar' -Arguments $tarArgs -WorkingDirectory $WorkingDirectory
}

function Convert-GoArch {
  param([string]$Arch)
  switch ($Arch.ToLowerInvariant()) {
    'x64' { return 'amd64' }
    'amd64' { return 'amd64' }
    'x86' { return '386' }
    '386' { return '386' }
    default { Fail "Unsupported backend arch: $Arch. Use x64 or x86." }
  }
}

function Convert-GoOS {
  param([string]$OS)
  switch ($OS.ToLowerInvariant()) {
    'linux' { return 'linux' }
    'windows' { return 'windows' }
    default { Fail "Unsupported backend OS: $OS. Use linux or windows." }
  }
}

function Build-BackendArtifact {
  param([hashtable]$Config)

  $backendDir = Resolve-RepoPath $Config.ProjectDir
  $stageDir = Join-Path $ArtifactRoot 'backend-stage'
  $archivePath = Join-Path $ArtifactRoot 'backend.tar.gz'
  $binaryDir = Join-Path $stageDir 'bin'
  $configSource = Resolve-RepoPath $Config.ConfigDir
  $runtimeAssetSource = $null

  if (-not (Test-Path $configSource)) {
    Fail "Backend config directory not found: $configSource"
  }
  if ($Config.RuntimeAssetsDir) {
    $runtimeAssetSource = Resolve-RepoPath $Config.RuntimeAssetsDir
    if (-not (Test-Path $runtimeAssetSource)) {
      Fail "Backend runtime assets directory not found: $runtimeAssetSource"
    }
  }

  New-CleanDirectory $stageDir
  New-Item -ItemType Directory -Path $binaryDir | Out-Null
  Copy-Item -Path $configSource -Destination (Join-Path $stageDir 'configs') -Recurse -Force

  $archiveItems = @('bin', 'configs')
  if ($runtimeAssetSource) {
    Copy-DirectoryContents -SourceDir $runtimeAssetSource -DestinationDir $stageDir -ExcludePaths @($configSource)
    $runtimeItems = Get-ChildItem -LiteralPath $runtimeAssetSource -Force |
      Where-Object { [System.IO.Path]::GetFullPath($_.FullName) -ne [System.IO.Path]::GetFullPath($configSource) } |
      ForEach-Object { $_.Name }
    if ($runtimeItems) {
      $archiveItems = @($archiveItems + $runtimeItems | Select-Object -Unique)
    }
  }

  $goos = Convert-GoOS $Config.BuildOS
  $goarch = Convert-GoArch $Config.BuildArch
  $binaryName = $Config.BinaryName
  if ($goos -eq 'windows' -and -not $binaryName.EndsWith('.exe')) {
    $binaryName += '.exe'
  }
  $binaryOutput = Join-Path $binaryDir $binaryName

  Write-Step "Building backend ($goos/$goarch)"
  $oldGoos = $env:GOOS
  $oldGoarch = $env:GOARCH
  try {
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    Invoke-Native -FilePath 'go' -Arguments @('build', '-o', $binaryOutput, $Config.Entry) -WorkingDirectory $backendDir
  }
  finally {
    $env:GOOS = $oldGoos
    $env:GOARCH = $oldGoarch
  }

  New-TarArchive -ArchivePath $archivePath -Items $archiveItems -WorkingDirectory $stageDir
  return @{
    ArchivePath = $archivePath
    BinaryName = $binaryName
  }
}

function Deploy-Backend {
  param(
    [hashtable]$RootConfig,
    [hashtable]$Artifact
  )

  $config = $RootConfig.Backend
  if ((Convert-GoOS $config.BuildOS) -ne 'linux') {
    Fail 'Backend remote start uses ssh + nohup, so Backend.BuildOS must be linux for automated deployment.'
  }

  $ssh = $RootConfig.Ssh
  $remoteArchive = "/tmp/$($config.BinaryName)-backend.tar.gz"
  $remoteDir = $config.RemoteDir
  $logFile = $config.LogFile
  $pidFile = $config.PidFile
  $logDir = Get-RemoteDirectoryName -Path $logFile
  $pidDir = Get-RemoteDirectoryName -Path $pidFile
  $processMatch = if ($config.ProcessMatch) { $config.ProcessMatch } else { $Artifact.BinaryName }
  $startCommand = if ($config.StartCommand) { $config.StartCommand } else { "./bin/$($Artifact.BinaryName)" }

  Write-Step 'Uploading backend package'
  Invoke-SshCommand -SshConfig $ssh -Command "mkdir -p $(Quote-Sh $remoteDir)"
  Copy-ToRemote -SshConfig $ssh -LocalPath $Artifact.ArchivePath -RemotePath $remoteArchive

  Write-Step 'Deploying backend on remote host'
  $command = @(
    "set -e",
    "mkdir -p $(Quote-Sh $remoteDir) $(Quote-Sh $logDir) $(Quote-Sh $pidDir)",
    "tar -xzf $(Quote-Sh $remoteArchive) -C $(Quote-Sh $remoteDir)",
    "rm -f $(Quote-Sh $remoteArchive)",
    "chmod +x $(Quote-Sh ($remoteDir.TrimEnd('/') + '/' + $startCommand.TrimStart('./').Replace('\\', '/')))",
    "if [ -f $(Quote-Sh $pidFile) ]; then kill `$(cat $(Quote-Sh $pidFile)) 2>/dev/null || true; rm -f $(Quote-Sh $pidFile); fi",
    "pkill -f $(Quote-Sh $processMatch) 2>/dev/null || true",
    "cd $(Quote-Sh $remoteDir)",
    "nohup $startCommand > $(Quote-Sh $logFile) 2>&1 < /dev/null & echo `$! > $(Quote-Sh $pidFile)"
  ) -join '; '
  Invoke-SshCommand -SshConfig $ssh -Command $command
}

function Build-FrontendArtifact {
  param(
    [hashtable]$ComponentConfig,
    [string]$ArtifactName
  )

  $projectDir = Resolve-RepoPath $ComponentConfig.ProjectDir
  $distDir = Join-Path $projectDir 'dist'
  $archivePath = Join-Path $ArtifactRoot $ArtifactName

  $buildEnv = @{}
  if ($ComponentConfig.ContainsKey('BuildEnv') -and $null -ne $ComponentConfig.BuildEnv) {
    $buildEnv = $ComponentConfig.BuildEnv
  }

  Write-Step "Building $($ComponentConfig.ProjectDir)"
  Invoke-WithEnvironmentOverrides -Variables $buildEnv -Action {
    Invoke-Native -FilePath 'npm.cmd' -Arguments @('run', 'build') -WorkingDirectory $projectDir
  }

  if (-not (Test-Path $distDir)) {
    Fail "Frontend dist directory not found: $distDir"
  }

  New-TarArchive -ArchivePath $archivePath -Items @('dist') -WorkingDirectory $projectDir
  return @{ ArchivePath = $archivePath }
}

function Deploy-Frontend {
  param(
    [hashtable]$RootConfig,
    [hashtable]$ComponentConfig,
    [hashtable]$Artifact,
    [string]$RemoteArchiveName,
    [string]$ComponentName
  )

  $ssh = $RootConfig.Ssh
  $remoteArchive = "/tmp/$RemoteArchiveName"
  $remoteDir = $ComponentConfig.RemoteDir

  Write-Step "Uploading $ComponentName package"
  Invoke-SshCommand -SshConfig $ssh -Command "mkdir -p $(Quote-Sh $remoteDir)"
  Copy-ToRemote -SshConfig $ssh -LocalPath $Artifact.ArchivePath -RemotePath $remoteArchive

  Write-Step "Deploying $ComponentName on remote host"
  $command = @(
    "set -e",
    "mkdir -p $(Quote-Sh $remoteDir)",
    "find $(Quote-Sh $remoteDir) -mindepth 1 -maxdepth 1 ! -name '.well-known' -exec rm -rf {} +",
    "tar -xzf $(Quote-Sh $remoteArchive) -C $(Quote-Sh $remoteDir) --strip-components 1",
    "rm -f $(Quote-Sh $remoteArchive)"
  ) -join '; '
  Invoke-SshCommand -SshConfig $ssh -Command $command
}

if (-not (Test-Path $ConfigPath)) {
  Fail "Config file not found: $ConfigPath"
}

$config = Import-PowerShellDataFile -Path $ConfigPath

Require-Command 'ssh'
Require-Command 'scp'
Require-Command 'tar'
if (-not $SkipBackend -and $config.Backend.Enabled) {
  Require-Command 'go'
}
if ((-not $SkipAdmin -and $config.AdminWeb.Enabled) -or (-not $SkipMerchantH5 -and $config.MerchantH5.Enabled)) {
  Require-Command 'npm.cmd'
}

New-Item -ItemType Directory -Path $ArtifactRoot -Force | Out-Null

$reloadNginx = $ForceNginxReload.IsPresent -or $config.Nginx.ReloadAfterDeploy

if (-not $SkipBackend -and $config.Backend.Enabled) {
  $backendArtifact = Build-BackendArtifact -Config $config.Backend
  Deploy-Backend -RootConfig $config -Artifact $backendArtifact
}

if (-not $SkipAdmin -and $config.AdminWeb.Enabled) {
  $adminArtifact = Build-FrontendArtifact -ComponentConfig $config.AdminWeb -ArtifactName 'admin-web.tar.gz'
  Deploy-Frontend -RootConfig $config -ComponentConfig $config.AdminWeb -Artifact $adminArtifact -RemoteArchiveName 'deskorder-admin-web.tar.gz' -ComponentName 'admin-web'
}

if (-not $SkipMerchantH5 -and $config.MerchantH5.Enabled) {
  $merchantArtifact = Build-FrontendArtifact -ComponentConfig $config.MerchantH5 -ArtifactName 'merchant-h5.tar.gz'
  Deploy-Frontend -RootConfig $config -ComponentConfig $config.MerchantH5 -Artifact $merchantArtifact -RemoteArchiveName 'deskorder-merchant-h5.tar.gz' -ComponentName 'merchant-h5'
}

if ($reloadNginx) {
  Write-Step 'Reloading nginx'
  Invoke-SshCommand -SshConfig $config.Ssh -Command $config.Nginx.ReloadCommand
}

Write-Step 'Deployment finished'