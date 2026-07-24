@echo off
setlocal
cd /d "%~dp0"

where powershell >nul 2>&1
if errorlevel 1 (
  echo [ERROR] PowerShell not found.
  pause
  exit /b 1
)

if not exist "%~dp0index.html" (
  echo [ERROR] Missing index.html next to this bat.
  echo Folder: %~dp0
  echo.
  echo Expected layout after unzip:
  echo   sales-front-windows\
  echo     start-sales.bat
  echo     serve.ps1
  echo     index.html
  echo     assets\
  echo.
  echo Do NOT copy only the .bat file. Unzip the whole zip.
  echo.
  dir /b "%~dp0"
  echo.
  pause
  exit /b 1
)

if not exist "%~dp0serve.ps1" (
  echo [ERROR] Missing serve.ps1 next to this bat.
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
