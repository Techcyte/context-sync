@echo off
REM Build the demo server into techcyte_context_sync_host.exe.
pushd "%~dp0"

go mod tidy
if errorlevel 1 goto :fail

go build -o techcyte_context_sync_host.exe cmd\tcs\main.go
if errorlevel 1 goto :fail

echo Built techcyte_context_sync_host.exe
popd
exit /b 0

:fail
echo Build failed.
popd
exit /b 1
