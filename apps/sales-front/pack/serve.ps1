# Serves static files next to this script, proxies /api → sales-manage :8083,
# and starts sales-manage.exe when present (Windows pack layout).
$ErrorActionPreference = 'Stop'

$port = 5175
$apiBase = 'http://127.0.0.1:8083'
$webRoot = $PSScriptRoot
$manageExe = Join-Path $webRoot 'sales-manage.exe'
$envFile = Join-Path $webRoot '.env'
$manageProc = $null

if (-not (Test-Path (Join-Path $webRoot 'index.html'))) {
  Write-Host '[ERROR] Missing index.html next to serve.ps1'
  Write-Host ("Folder: {0}" -f $webRoot)
  exit 1
}

function Import-DotEnv([string]$path) {
  if (-not (Test-Path -LiteralPath $path)) { return }
  Get-Content -LiteralPath $path | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq '' -or $line.StartsWith('#')) { return }
    $idx = $line.IndexOf('=')
    if ($idx -lt 1) { return }
    $key = $line.Substring(0, $idx).Trim()
    $val = $line.Substring($idx + 1).Trim()
    [Environment]::SetEnvironmentVariable($key, $val, 'Process')
  }
}

function Wait-ApiHealthy([string]$base, [int]$tries = 40) {
  for ($i = 0; $i -lt $tries; $i++) {
    try {
      $r = Invoke-WebRequest -Uri ($base + '/health') -UseBasicParsing -TimeoutSec 1
      if ($r.StatusCode -eq 200) { return $true }
    } catch {
      Start-Sleep -Milliseconds 400
    }
  }
  return $false
}

function Proxy-ToApi($ctx) {
  Add-Type -AssemblyName System.Net.Http | Out-Null
  $client = New-Object System.Net.Http.HttpClient
  $client.Timeout = [TimeSpan]::FromSeconds(60)
  try {
    $uri = $apiBase + $ctx.Request.RawUrl
    $method = New-Object System.Net.Http.HttpMethod $ctx.Request.HttpMethod
    $msg = New-Object System.Net.Http.HttpRequestMessage $method, $uri

    if ($ctx.Request.HasEntityBody) {
      $buf = New-Object byte[] $ctx.Request.ContentLength64
      $read = 0
      while ($read -lt $buf.Length) {
        $n = $ctx.Request.InputStream.Read($buf, $read, $buf.Length - $read)
        if ($n -le 0) { break }
        $read += $n
      }
      if ($read -lt $buf.Length) {
        $tmp = New-Object byte[] $read
        [Array]::Copy($buf, $tmp, $read)
        $buf = $tmp
      }
      $msg.Content = New-Object System.Net.Http.ByteArrayContent $buf
      if ($ctx.Request.ContentType) {
        $msg.Content.Headers.TryAddWithoutValidation('Content-Type', $ctx.Request.ContentType) | Out-Null
      }
    }

    $resp = $client.SendAsync($msg).GetAwaiter().GetResult()
    $bytes = $resp.Content.ReadAsByteArrayAsync().GetAwaiter().GetResult()
    $ctx.Response.StatusCode = [int]$resp.StatusCode
    if ($resp.Content.Headers.ContentType) {
      $ctx.Response.ContentType = $resp.Content.Headers.ContentType.ToString()
    } else {
      $ctx.Response.ContentType = 'application/json; charset=utf-8'
    }
    $ctx.Response.ContentLength64 = $bytes.Length
    if ($bytes.Length -gt 0) {
      $ctx.Response.OutputStream.Write($bytes, 0, $bytes.Length)
    }
  } finally {
    $client.Dispose()
  }
}

# Start sales-manage when packaged next to this script
if (Test-Path -LiteralPath $manageExe) {
  Import-DotEnv $envFile
  Write-Host 'Starting sales-manage.exe ...'
  $manageProc = Start-Process -FilePath $manageExe -WorkingDirectory $webRoot -PassThru -WindowStyle Minimized
  if (-not (Wait-ApiHealthy $apiBase)) {
    Write-Host '[ERROR] sales-manage did not become healthy on :8083'
    Write-Host 'Check MySQL is running, database sales_manage exists, and .env DB_* settings.'
    if ($null -ne $manageProc -and -not $manageProc.HasExited) {
      Stop-Process -Id $manageProc.Id -Force -ErrorAction SilentlyContinue
    }
    exit 1
  }
  Write-Host 'sales-manage is up (http://127.0.0.1:8083)'
} else {
  Write-Host '[WARN] sales-manage.exe not found; /api will fail unless something else listens on :8083'
}

$mime = @{
  '.html' = 'text/html; charset=utf-8'
  '.js'   = 'application/javascript; charset=utf-8'
  '.css'  = 'text/css; charset=utf-8'
  '.json' = 'application/json; charset=utf-8'
  '.svg'  = 'image/svg+xml'
  '.png'  = 'image/png'
  '.jpg'  = 'image/jpeg'
  '.jpeg' = 'image/jpeg'
  '.ico'  = 'image/x-icon'
  '.woff' = 'font/woff'
  '.woff2'= 'font/woff2'
  '.ttf'  = 'font/ttf'
  '.map'  = 'application/json'
  '.bat'  = 'text/plain'
  '.ps1'  = 'text/plain'
  '.txt'  = 'text/plain; charset=utf-8'
  '.env'  = 'text/plain'
  '.exe'  = 'application/octet-stream'
}

$prefix = "http://127.0.0.1:$port/"
$listener = New-Object System.Net.HttpListener
$listener.Prefixes.Add($prefix)

try {
  $listener.Start()
} catch {
  Write-Host "[ERROR] Cannot listen on $prefix (port in use?)"
  Write-Host $_.Exception.Message
  if ($null -ne $manageProc -and -not $manageProc.HasExited) {
    Stop-Process -Id $manageProc.Id -Force -ErrorAction SilentlyContinue
  }
  exit 1
}

Write-Host "sales-front running: $prefix"
Write-Host 'Close this window to stop (also stops sales-manage).'
Start-Process $prefix

function Get-ContentType([string]$ext) {
  if ($mime.ContainsKey($ext)) { return $mime[$ext] }
  return 'application/octet-stream'
}

try {
  while ($listener.IsListening) {
    $ctx = $listener.GetContext()
    $req = $ctx.Request
    $res = $ctx.Response
    try {
      $path = $req.Url.AbsolutePath
      if ($path.StartsWith('/api/', [StringComparison]::OrdinalIgnoreCase) -or
          $path.Equals('/api', [StringComparison]::OrdinalIgnoreCase)) {
        Proxy-ToApi $ctx
        continue
      }

      $rel = [Uri]::UnescapeDataString($path.TrimStart('/'))
      if ([string]::IsNullOrWhiteSpace($rel)) { $rel = 'index.html' }

      # Do not serve secrets / binaries as static downloads by accident via SPA
      if ($rel -in @('.env', 'sales-manage.exe', 'serve.ps1', 'start-sales.bat')) {
        $res.StatusCode = 404
        $res.Close()
        continue
      }

      $candidate = [System.IO.Path]::GetFullPath((Join-Path $webRoot ($rel -replace '/', [IO.Path]::DirectorySeparatorChar)))
      $rootFull = [System.IO.Path]::GetFullPath($webRoot)
      if (-not $candidate.StartsWith($rootFull, [StringComparison]::OrdinalIgnoreCase)) {
        $res.StatusCode = 403
        $res.Close()
        continue
      }

      if (-not (Test-Path -LiteralPath $candidate -PathType Leaf)) {
        $candidate = Join-Path $webRoot 'index.html'
      }

      $bytes = [System.IO.File]::ReadAllBytes($candidate)
      $ext = [System.IO.Path]::GetExtension($candidate).ToLowerInvariant()
      $res.ContentType = Get-ContentType $ext
      $res.ContentLength64 = $bytes.Length
      $res.StatusCode = 200
      $res.OutputStream.Write($bytes, 0, $bytes.Length)
    } catch {
      try { $res.StatusCode = 500 } catch {}
    } finally {
      try { $res.OutputStream.Close() } catch {}
      try { $res.Close() } catch {}
    }
  }
} finally {
  if ($listener.IsListening) { $listener.Stop() }
  $listener.Close()
  if ($null -ne $manageProc -and -not $manageProc.HasExited) {
    Write-Host 'Stopping sales-manage...'
    Stop-Process -Id $manageProc.Id -Force -ErrorAction SilentlyContinue
  }
}
