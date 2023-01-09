windows_exec_file=windows-generator.exe
linux_exec_file=linux-generator

all: windows linux

windows:
	GOOS=windows CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o $(windows_exec_file) 

windows-exec: windows
	./$(windows_exec_file)

linux:
	go build -o $(linux_exec_file)

linux-exec: linux
	./$(linux_exec_file)

test:
	go test -v *.go
