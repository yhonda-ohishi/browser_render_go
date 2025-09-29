@echo off
echo Generating Protocol Buffers code...

:: Check if buf is installed
where buf >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo buf is not installed. Please run scripts\install-buf.bat first
    exit /b 1
)

:: Install required Go plugins
echo Installing/updating Go protoc plugins...
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

:: Generate code
echo Running buf generate...
buf generate

if %ERRORLEVEL% EQU 0 (
    echo Protocol Buffers generation completed!
    echo Generated files are in the gen\ directory
) else (
    echo Failed to generate Protocol Buffers
    exit /b 1
)