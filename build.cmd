set GOOS=linux
set GOARCH=amd64
echo 'listener_%GOOS%_%GOARCH%'���o�͂��Ă��܂�
go build -o "back_%GOOS%_%GOARCH%" .
pause