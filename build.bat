SET dir=%HOMEDRIVE%%HOMEPATH%\.ankat
SET exe=%dir%\ankathost.exe
buildmingw32 go build -o %exe% -ldflags="-H windowsgui" github.com/fpawel/ankat/cmd
go build -o %dir%\runankat.exe -ldflags="-H windowsgui" "github.com/fpawel/ankat/run"
start %dir%
