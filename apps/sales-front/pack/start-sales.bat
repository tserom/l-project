@param off
setlocal
cd /d "%~dp0"

where powershell >nul 2>&1
if errorlevel 1 (
  echo [错误] 未找到 PowerShell，无法启动。
  pause
  exit /b 1
)

if not exist "%~dp0web\index.html" (
  echo [错误] 未找到 web\index.html，请确认解压完整。
  pause
  exit /b 1
)

echo 正在启动销售单（本机 http://127.0.0.1:5175 ）...
echo 关闭本窗口即停止服务。
echo.

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0serve.ps1"
set ERR=%ERRORLEVEL%
if not "%ERR%"=="0" (
  echo.
  echo [错误] 启动失败，退出码 %ERR%
  pause
)
exit /b %ERR%
