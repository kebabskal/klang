@echo off
if "%~1"=="" (
    echo usage: run.bat ^<file.k^>
    exit /b 1
)
"c:/Program Files\Go\bin\go.exe" build -o build\kl.exe .\cmd\kl\ && build\kl.exe dev %*
