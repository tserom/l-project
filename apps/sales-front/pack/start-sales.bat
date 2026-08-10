@echo off
setlocal EnableDelayedExpansion
set "ROOT=%~dp0"
cd /d "!ROOT!"

where powershell >nul 2>&1
if errorlevel 1 goto :err_no_powershell

if not exist "!ROOT!index.html" goto :err_no_index
if not exist "!ROOT!serve.ps1" goto :err_no_serve

echo Starting sales-front at http://127.0.0.1:5175 ...
echo Close this window to stop.
echo.

powershell.exe -NoProfile -ExecutionPolicy Bypass -File "!ROOT!serve.ps1"
set ERR=!ERRORLEVEL!
if not "!ERR!"=="0" (
  echo.
  echo [ERROR] Start failed, exit code !ERR!
  pause
)
exit /b !ERR!

:err_no_powershell
echo [ERROR] PowerShell not found.
pause
exit /b 1

:err_no_index
echo [ERROR] Missing index.html next to this bat.
echo Folder: !ROOT!
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
dir /b "!ROOT!"
echo.
pause
exit /b 1

:err_no_serve
echo [ERROR] Missing serve.ps1 next to this bat.
pause
exit /b 1
