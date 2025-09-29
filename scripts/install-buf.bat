@echo off
echo Installing buf for Windows...

:: Download buf.exe
curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Windows-x86_64.exe" -o buf.exe

:: Check if download was successful
if exist buf.exe (
    echo buf.exe downloaded successfully!
    echo Please move buf.exe to a directory in your PATH
    echo For example: move buf.exe C:\Windows\System32\
    echo.
    echo After moving, you can run: buf --version
) else (
    echo Failed to download buf.exe
    exit /b 1
)