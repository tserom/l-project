@echo off
setlocal
cd /d "%~dp0"

where powershell >nul 2>&1
if errorlevel 1 (
  echo [ERROR] PowerShell not found.
  pause
  exit /b 1
)

if not exist "%~dp0web\index.html" (
  echo [ERROR] Missing web\index.html - incomplete unzip?
  pause
  exit /b 1
)

echo Starting sales-front at http://127.0.0.1:5175 ...
echo Close this window to stop.
echo.

powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0serve.ps1"
set ERR=%ERRORLEVEL%
if not "%ERR%"=="0" (
  echo.
  echo [ERROR] Start failed, exit code %ERR%
  pause
)
exit /b %ERR%
