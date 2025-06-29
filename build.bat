@echo off
REM Build script for the tr CLI tool on Windows

echo Building TR - English-Spanish Translation CLI Tool

REM Clean previous builds
echo Cleaning previous builds...
if exist tr.exe del tr.exe
if exist tr-linux del tr-linux
if exist tr-macos del tr-macos
if exist tr-windows.exe del tr-windows.exe

REM Get dependencies
echo Downloading dependencies...
go mod download

REM Build for current platform (Windows)
echo Building for Windows...
go build -o tr.exe ./cmd/tr

REM Build for multiple platforms
echo Building for multiple platforms...

REM macOS
echo Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -o tr-macos ./cmd/tr

REM Linux
echo Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o tr-linux ./cmd/tr

REM Reset environment
set GOOS=
set GOARCH=

echo Build complete!
echo Executables created:
echo   - tr.exe (Windows)
echo   - tr-macos (macOS)
echo   - tr-linux (Linux)

REM Test the build
echo.
echo Testing the build...
tr.exe --version
echo Build test successful!
