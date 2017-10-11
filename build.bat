set GOPATH=D:\Documents\Projets\sandbox-go;D:\Documents\Projets\display_monitor;D:\Donnees\Go

go build -o generated/task_server.exe src/main/task_manager_server.go
go build -o generated/node_server_x64.exe src/main/node_server.go

set GOOS=windows
set GOARCH=386

go build -o generated/node_server_x86.exe src/main/node_server.go
cp resources/launch_task.html generated

set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -o generated/node_server_arm64 src/main/node_server.go

