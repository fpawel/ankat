SET dir=%HOMEDRIVE%%HOMEPATH%\.ankat
copy %GOPATH%\bin\pp.exe %dir%\pp.exe
SET exe=%dir%\ankathost.exe
SET buildExe=%Temp%\buildankat32.exe
if exist %exe% (del %exe%)
go build -o %buildExe% "github.com/fpawel/ankat/build"
%buildExe% go build -ldflags="-H windowsgui" -o %exe%
del %buildExe%
start %dir%
