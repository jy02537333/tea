Param(
    [string]$RootPath = (Get-Location).Path,
    [int]$AdminPort = 8081,
    [int]$WxPort = 8082,
    [int]$TeaApiHealthTimeout = 30
)

Set-StrictMode -Version Latest

function Ensure-Dir([string]$p) {
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p | Out-Null }
}

function Is-CmdAvailable([string]$cmd) {
    return (Get-Command $cmd -ErrorAction SilentlyContinue) -ne $null
}

function Start-Background([string]$name, [string]$exe, [array]$args, [string]$cwd, [string]$outLog, [string]$errLog) {
    Write-Output "Starting $name..."
    Ensure-Dir (Split-Path $outLog)
    Ensure-Dir (Split-Path $errLog)
    # Build a single argument string to avoid null-element validation issues
    $arguments = ""
    if ($null -ne $args -and $args.Count -gt 0) {
        $arguments = ($args -join ' ')
    }
    # Start-Process will run independently (background)
    if ($arguments -ne "") {
        Start-Process -FilePath $exe -ArgumentList $arguments -WorkingDirectory $cwd -RedirectStandardOutput $outLog -RedirectStandardError $errLog -WindowStyle Hidden | Out-Null
    } else {
        Start-Process -FilePath $exe -WorkingDirectory $cwd -RedirectStandardOutput $outLog -RedirectStandardError $errLog -WindowStyle Hidden | Out-Null
    }
    Write-Output "$name started (logs: $outLog, $errLog)"
}

function Wait-For-Http([string]$url, [int]$timeoutSeconds) {
    $sw = [Diagnostics.Stopwatch]::StartNew()
    while ($sw.Elapsed.TotalSeconds -lt $timeoutSeconds) {
        try {
            $r = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
            if ($r.StatusCode -ge 200 -and $r.StatusCode -lt 400) { return $true }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    return $false
}

Push-Location $RootPath
Ensure-Dir "$RootPath\logs"

# 1) Start tea-api (Go)
$teaDir = Join-Path $RootPath 'tea-api'
$teaOut = Join-Path $RootPath 'logs\tea-api.out.log'
$teaErr = Join-Path $RootPath 'logs\tea-api.err.log'
if (-not (Test-Path $teaDir)) { Write-Error "tea-api directory not found: $teaDir"; Pop-Location; exit 1 }

Start-Background 'tea-api' 'go' @('run','main.go') $teaDir $teaOut $teaErr

Write-Output 'Waiting for tea-api health...'
if (Wait-For-Http 'http://localhost:8080/api/v1/health' $TeaApiHealthTimeout) {
    Write-Output 'tea-api healthy.'
} else {
    Write-Warning "tea-api did not become healthy within ${TeaApiHealthTimeout}s. Check logs: $teaOut / $teaErr"
}

# 2) Start Admin-FE static server
$adminDir = Join-Path $RootPath 'Admin-FE'
$adminOut = Join-Path $RootPath "logs\admin-fe.out.log"
$adminErr = Join-Path $RootPath "logs\admin-fe.err.log"
if (Test-Path $adminDir) {
    if (Is-CmdAvailable 'npx') {
        Start-Background 'admin-fe (http-server via npx)' 'npx' @('http-server','-p',"$AdminPort",'.') $adminDir $adminOut $adminErr
    } elseif (Is-CmdAvailable 'python') {
        Start-Background 'admin-fe (python -m http.server)' 'python' @('-m','http.server',"$AdminPort") $adminDir $adminOut $adminErr
    } else {
        Write-Warning 'No npx or python found. Admin-FE not started. Please install one or serve files manually.'
    }
} else { Write-Warning "Admin-FE dir not found: $adminDir" }

# 3) Start WX-FE static server
$wxDir = Join-Path $RootPath 'WX-FE'
$wxOut = Join-Path $RootPath "logs\wx-fe.out.log"
$wxErr = Join-Path $RootPath "logs\wx-fe.err.log"
if (Test-Path $wxDir) {
    if (Is-CmdAvailable 'npx') {
        Start-Background 'wx-fe (http-server via npx)' 'npx' @('http-server','-p',"$WxPort",'.') $wxDir $wxOut $wxErr
    } elseif (Is-CmdAvailable 'python') {
        Start-Background 'wx-fe (python -m http.server)' 'python' @('-m','http.server',"$WxPort") $wxDir $wxOut $wxErr
    } else {
        Write-Warning 'No npx or python found. WX-FE not started. Please install one or serve files manually.'
    }
} else { Write-Warning "WX-FE dir not found: $wxDir" }

# 4) Verify Admin-FE and WX-FE if started
if (Test-Path $adminDir) {
    $adminUrl = "http://localhost:$AdminPort/"
    Write-Output "Checking Admin-FE at $adminUrl"
    if (Wait-For-Http $adminUrl 10) { Write-Output 'Admin-FE reachable.' } else { Write-Warning 'Admin-FE not reachable (check logs).' }
}
if (Test-Path $wxDir) {
    $wxUrl = "http://localhost:$WxPort/"
    Write-Output "Checking WX-FE at $wxUrl"
    if (Wait-For-Http $wxUrl 10) { Write-Output 'WX-FE reachable.' } else { Write-Warning 'WX-FE not reachable (check logs).' }
}

Write-Output 'All start commands issued. Use logs in the `logs` folder to inspect output.'
Pop-Location
