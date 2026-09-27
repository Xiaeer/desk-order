[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$scriptRepoRoot = Split-Path -Parent $PSScriptRoot
$scriptRepoRootForGit = $scriptRepoRoot -replace '\\', '/'
$repoRootOutput = @(& git -c "safe.directory=$scriptRepoRootForGit" -C $scriptRepoRoot rev-parse --show-toplevel 2>$null)
$gitExitCode = $LASTEXITCODE
$repoRoot = $repoRootOutput | Select-Object -First 1
if ($gitExitCode -ne 0 -or [string]::IsNullOrWhiteSpace($repoRoot)) {
    Write-Error 'This script must be run from inside a Git repository.'
    exit 2
}

$repoRoot = $repoRoot.Trim()
$repoRootForGit = $repoRoot -replace '\\', '/'
$findings = [System.Collections.Generic.List[string]]::new()

function Add-Finding {
    param(
        [Parameter(Mandatory)] [string] $Rule,
        [Parameter(Mandatory)] [string] $Path,
        [int] $Line = 0
    )

    $location = if ($Line -gt 0) { "${Path}:${Line}" } else { $Path }
    $findings.Add("[$Rule] $location")
}

function Test-PlaceholderValue {
    param([AllowEmptyString()] [string] $Value)

    $candidate = $Value.Trim().TrimEnd(',', ';').Trim()
    if ($candidate.Length -ge 2) {
        $first = $candidate[0]
        $last = $candidate[$candidate.Length - 1]
        if (($first -eq '"' -and $last -eq '"') -or ($first -eq "'" -and $last -eq "'")) {
            $candidate = $candidate.Substring(1, $candidate.Length - 2).Trim()
        }
    }

    if ([string]::IsNullOrWhiteSpace($candidate)) { return $true }
    if ($candidate -match '^(?i:null|nil|none|false|0)$') { return $true }
    if ($candidate -match '^\$\{[^}]+\}$') { return $true }
    if ($candidate -match '^\$env:[A-Za-z_][A-Za-z0-9_]*$') { return $true }
    if ($candidate -match '(?i)(process\.env|os\.Getenv|getenv\s*\()') { return $true }
    if ($candidate -match '(?i)(example|sample|dummy|placeholder|replace|change[-_ ]?me|your[-_ ]|test[-_ ]?(only|value)?|touristappid|x{4,}|<[^>]+>)') { return $true }

    return $false
}

Push-Location -LiteralPath $repoRoot
try {
    $fileList = @(& git -c "safe.directory=$repoRootForGit" ls-files --cached --others --exclude-standard)
    if ($LASTEXITCODE -ne 0) {
        Write-Error 'Unable to enumerate repository files.'
        exit 2
    }

    $pathRules = @(
        @{ Name = 'local-env-file'; Pattern = '(^|/)\.env(?:\.[^/]+)?$'; Allow = '(?i)\.example$' },
        @{ Name = 'local-runtime-config'; Pattern = '(^|/)config\.local\.js$'; Allow = '$a' },
        @{ Name = 'wechat-private-config'; Pattern = '(^|/)project\.private\.config\.json$'; Allow = '$a' },
        @{ Name = 'backend-real-config'; Pattern = '^backend/configs/config\.ya?ml$'; Allow = '$a' },
        @{ Name = 'deploy-real-config'; Pattern = '^deploy/env/[^/]+/backend-configs/config\.ya?ml$'; Allow = '$a' },
        @{ Name = 'deploy-private-profile'; Pattern = '^deploy/(?:profiles/)?config(?:\.[^/]+)?\.psd1$'; Allow = '(?i)config\.example\.psd1$' },
        @{ Name = 'private-key-or-certificate'; Pattern = '(?i)\.(?:pem|key|p12|pfx|jks|keystore|crt|cer|der)$'; Allow = '$a' },
        @{ Name = 'database-or-backup-artifact'; Pattern = '(?i)\.(?:dump|bak|backup)$'; Allow = '$a' }
    )

    $tokenRules = @(
        @{ Name = 'private-key-block'; Pattern = '-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----' },
        @{ Name = 'aws-access-key'; Pattern = '\b(?:AKIA|ASIA)[0-9A-Z]{16}\b' },
        @{ Name = 'github-token'; Pattern = '\b(?:gh[pousr]_[A-Za-z0-9]{36,255}|github_pat_[A-Za-z0-9_]{20,255})\b' },
        @{ Name = 'google-api-key'; Pattern = '\bAIza[0-9A-Za-z_-]{35}\b' },
        @{ Name = 'slack-token'; Pattern = '\bxox[baprs]-[A-Za-z0-9-]{10,}\b' },
        @{ Name = 'stripe-live-secret'; Pattern = '\bsk_live_[A-Za-z0-9]{16,}\b' },
        @{ Name = 'wechat-app-id'; Pattern = '\bwx[0-9a-fA-F]{16}\b' }
    )

    $binaryExtensions = @(
        '.7z', '.avi', '.bmp', '.class', '.dll', '.doc', '.docx', '.exe', '.gif', '.gz',
        '.ico', '.jar', '.jpeg', '.jpg', '.mov', '.mp3', '.mp4', '.pdf', '.png', '.so',
        '.tar', '.tif', '.tiff', '.ttf', '.wav', '.webm', '.woff', '.woff2', '.xls', '.xlsx', '.zip'
    )

    foreach ($entry in $fileList) {
        if ([string]::IsNullOrWhiteSpace($entry)) { continue }

        $relativePath = $entry.Trim() -replace '\\', '/'

        foreach ($rule in $pathRules) {
            if ($relativePath -match $rule.Pattern -and $relativePath -notmatch $rule.Allow) {
                Add-Finding -Rule $rule.Name -Path $relativePath
            }
        }

        $fullPath = Join-Path -Path $repoRoot -ChildPath ($relativePath -replace '/', [IO.Path]::DirectorySeparatorChar)
        if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf)) { continue }
        if ($binaryExtensions -contains ([IO.Path]::GetExtension($relativePath).ToLowerInvariant())) { continue }

        try {
            $bytes = [IO.File]::ReadAllBytes($fullPath)
            if ($bytes -contains 0) { continue }
            $content = [Text.Encoding]::UTF8.GetString($bytes)
        }
        catch {
            Add-Finding -Rule 'unreadable-file' -Path $relativePath
            continue
        }

        # Avoid detecting the scanner's own rule definitions.
        if ($relativePath -ne 'scripts/check-secrets.ps1') {
            foreach ($rule in $tokenRules) {
                $matches = [regex]::Matches($content, $rule.Pattern)
                foreach ($match in $matches) {
                    $line = 1 + (($content.Substring(0, $match.Index) -split "`n").Count - 1)
                    Add-Finding -Rule $rule.Name -Path $relativePath -Line $line
                }
            }

            $credentialUrlPattern = '(?i)\b(?:mysql|postgres(?:ql)?|mongodb(?:\+srv)?|redis|amqps?|https?)://[^/\s:@]+:(?<password>[^@\s/]+)@'
            foreach ($match in [regex]::Matches($content, $credentialUrlPattern)) {
                if (-not (Test-PlaceholderValue -Value $match.Groups['password'].Value)) {
                    $line = 1 + (($content.Substring(0, $match.Index) -split "`n").Count - 1)
                    Add-Finding -Rule 'credential-in-url' -Path $relativePath -Line $line
                }
            }
        }

        $isConfigLike = $relativePath -match '(?i)(^|/)(?:[^/]*(?:config|settings)[^/]*\.(?:json|ya?ml|toml|ini|properties|psd1)|\.env(?:\.[^/]+)?)$'
        if (-not $isConfigLike) { continue }

        $lineNumber = 0
        foreach ($lineText in ($content -split "`r?`n")) {
            $lineNumber++
            if ($lineText -match '^\s*(?:["'']?)(?<key>[A-Za-z0-9_.-]+)(?:["'']?)\s*[:=]\s*(?<value>.+?)\s*$') {
                $key = $Matches['key']
                $value = $Matches['value']
                $secretKeyPattern = '(?i)(?:^|[_.-])(?:secret|password|passwd|token|api[-_.]?key|app[-_.]?secret|private[-_.]?key|access[-_.]?key|jwt[-_.]?(?:secret|key)|merchant[-_.]?key|mch[-_.]?key|signing[-_.]?key)(?:$|[_.-])'
                if ($key -match $secretKeyPattern -and -not (Test-PlaceholderValue -Value $value)) {
                    Add-Finding -Rule 'hardcoded-config-secret' -Path $relativePath -Line $lineNumber
                }
            }
        }
    }
}
finally {
    Pop-Location
}

$uniqueFindings = @($findings | Sort-Object -Unique)
if ($uniqueFindings.Count -gt 0) {
    Write-Host 'Secret policy check failed. Values are intentionally not displayed.' -ForegroundColor Red
    $uniqueFindings | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
    exit 1
}

Write-Host "Secret policy check passed ($($fileList.Count) repository files checked)." -ForegroundColor Green
