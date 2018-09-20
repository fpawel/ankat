SET dir=%HOMEDRIVE%%HOMEPATH%\.ankat
buildmingw32 go build -o %dir%\ankathost.exe -ldflags="-H windowsgui" github.com/fpawel/ankat/cmd
go build -o %dir%\runankat.exe -ldflags="-H windowsgui" github.com/fpawel/ankat/run
start %dir%
