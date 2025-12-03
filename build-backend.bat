@echo off
REM Build script for Audio Track Switcher (Windows)
REM This script builds the Go backend for Windows

echo Building Go backend for Windows...

cd go-backend

echo Building executable...
go build -o ..\src-tauri\audio-track-backend.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo Go backend build complete!
    echo Binary created: src-tauri\audio-track-backend.exe
) else (
    echo Build failed!
    exit /b 1
)

cd ..
