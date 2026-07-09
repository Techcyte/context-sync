@echo off
REM Build and run the demo server.
pushd "%~dp0"

call "%~dp0build.bat"
if errorlevel 1 (
    popd
    exit /b 1
)

techcyte_context_sync_host.exe
popd
