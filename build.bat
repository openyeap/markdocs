@echo off
setlocal

FOR /F "delims=" %%i IN ("%cd%") DO (
    set name=%%~ni
) 

if "%1" == "" (
    del go.mod
    del go.sum
    go mod init fdsa.ltd/%name%
    go mod tidy
)

REM gofmt -w src
set GOOS=linux
set GOARCH=amd64

if "%1" == "" (
    go build -o ./bin/%name% fdsa.ltd/%name%/src
) else (
    go build -ldflags="-s -w" -o ./bin/%name% fdsa.ltd/%name%/src
)
echo go: linux version is finished ok

set GOOS=windows
set GOARCH=amd64
if "%1" == "" (
    go build -o ./bin/%name%.exe fdsa.ltd/%name%/src
) else (
    go build -ldflags="-s -w" -H=windowsgui -o ./bin/%name%.exe fdsa.ltd/%name%/src
)
echo go: windows version is finished ok


if "%1" == "" (
    echo go: package...
) else (
    upx -9 ./bin/%name%.exe
    upx -9 ./bin/%name%
)

if not exist target (
    mkdir target
)
cd ..

tar -czvf %name%/target/%name%.tar.gz %name%/install %name%/install.bat %name%/bin %name%/*.md %name%/docs
zip %name%/target/%name%.zip %name%/install %name%/install.bat %name%/bin %name%/*.md %name%/docs
echo go: all is finished
