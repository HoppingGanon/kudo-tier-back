set GOOS=linux
set GOARCH=amd64
echo 'listener_%GOOS%_%GOARCH%'を出力しています
go build -o "back_%GOOS%_%GOARCH%" .
pause