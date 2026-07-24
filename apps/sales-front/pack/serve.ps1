# Serves files next to this script (flat pack layout) with SPA fallback.
$ErrorActionPreference = 'Stop'

$port = 5175
$webRoot = $PSScriptRoot
if (-not (Test-Path (Join-Path $webRoot 'index.html'))) {
  Write-Host '[ERROR] Missing index.html next to serve.ps1'
  Write-Host ("Folder: {0}" -f $webRoot)
  exit 1
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
}

$prefix = "http://127.0.0.1:$port/"
$listener = New-Object System.Net.HttpListener
$listener.Prefixes.Add($prefix)

try {
  $listener.Start()
} catch {
  Write-Host "[ERROR] Cannot listen on $prefix (port in use?)"
  Write-Host $_.Exception.Message
  exit 1
}

Write-Host "sales-front running: $prefix"
Write-Host 'Close this window to stop.'
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
      $rel = [Uri]::UnescapeDataString($req.Url.AbsolutePath.TrimStart('/'))
      if ([string]::IsNullOrWhiteSpace($rel)) { $rel = 'index.html' }

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
}
