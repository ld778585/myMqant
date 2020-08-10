@echo off
set GOPATH=%~dp0
SET GOOS=windows
echo %GOPATH%
go build -o ./bin/gameServer.exe main
XCOPY .\bin\game.exe ..\build /y
XCOPY .\bin\conf ..\build\conf\ /y