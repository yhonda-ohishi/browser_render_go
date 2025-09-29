@echo off
echo Generating Protocol Buffers code for Go...

:: Install required Go plugins
echo Installing Go protoc plugins...
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

:: Download googleapis proto files (needed for annotations)
echo Downloading googleapis proto files...
if not exist "third_party\googleapis" (
    mkdir third_party\googleapis
    git clone --depth=1 https://github.com/googleapis/googleapis.git third_party\googleapis
)

:: Generate Go code using protoc directly
echo Generating Go code...
protoc ^
    --proto_path=proto ^
    --proto_path=third_party\googleapis ^
    --go_out=gen\proto --go_opt=paths=source_relative ^
    --go-grpc_out=gen\proto --go-grpc_opt=paths=source_relative ^
    --grpc-gateway_out=gen\proto --grpc-gateway_opt=paths=source_relative ^
    proto\browser_render\v1\browser_render.proto

if %ERRORLEVEL% EQU 0 (
    echo Protocol Buffers generation completed!
    echo Generated files are in the gen\proto directory
) else (
    echo Failed to generate Protocol Buffers
    echo Make sure protoc is installed: https://github.com/protocolbuffers/protobuf/releases
    exit /b 1
)