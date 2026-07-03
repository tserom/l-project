@echo off
setlocal EnableDelayedExpansion

cd /d "%~dp0..\.."
set "ROOT=%CD%"

set "CENTER_PID="
set "MANAGE_PID="

if exist "%ROOT%\bin\stock-center.exe" (
  echo Starting bin\stock-center...
  start "stock-center" /B cmd /c "cd /d %ROOT%\apps\stock-center && if exist .env for /f \"usebackq tokens=1,* delims==\" %%a in (.env) do set %%a=%%b && %ROOT%\bin\stock-center.exe"
) else if exist "%ROOT%\bin\stock-center" (
  echo Starting bin\stock-center...
  start "stock-center" /B cmd /c "cd /d %ROOT%\apps\stock-center && if exist .env for /f \"usebackq tokens=1,* delims==\" %%a in (.env) do set %%a=%%b && %ROOT%\bin\stock-center"
) else (
  echo bin\stock-center not found; using go run...
  start "stock-center" /B cmd /c "cd /d %ROOT%\apps\stock-center && go run ./cmd/server"
)

call :wait_for_health http://localhost:8081/health stock-center
if errorlevel 1 exit /b 1

if exist "%ROOT%\bin\stock-manage.exe" (
  echo Starting bin\stock-manage...
  start "stock-manage" /B cmd /c "cd /d %ROOT%\apps\stock-manage && if exist .env for /f \"usebackq tokens=1,* delims==\" %%a in (.env) do set %%a=%%b && %ROOT%\bin\stock-manage.exe"
) else if exist "%ROOT%\bin\stock-manage" (
  echo Starting bin\stock-manage...
  start "stock-manage" /B cmd /c "cd /d %ROOT%\apps\stock-manage && if exist .env for /f \"usebackq tokens=1,* delims==\" %%a in (.env) do set %%a=%%b && %ROOT%\bin\stock-manage"
) else (
  echo bin\stock-manage not found; using go run...
  start "stock-manage" /B cmd /c "cd /d %ROOT%\apps\stock-manage && go run ./cmd/server"
)

call :wait_for_health http://localhost:8082/health stock-manage
if errorlevel 1 exit /b 1

echo Opening http://localhost:8082 ...
start http://localhost:8082

echo Services running. Close this window or press Ctrl+C to stop background jobs.
pause
exit /b 0

:wait_for_health
set "URL=%~1"
set "NAME=%~2"
set /a COUNT=0
:health_loop
curl -sf "%URL%" >nul 2>&1
if not errorlevel 1 (
  echo %NAME% is healthy
  exit /b 0
)
timeout /t 1 /nobreak >nul
set /a COUNT+=1
if !COUNT! geq 60 (
  echo ERROR: %NAME% did not become healthy at %URL%
  echo Check MySQL is running and .env files are configured.
  exit /b 1
)
goto health_loop
