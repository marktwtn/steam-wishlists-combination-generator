all: windows linux

windows:
	GOOS=windows CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o windows-generator.exe

windows-exec: windows
	./windows-generator.exe

linux:
	go build -o linux-generator

linux-exec: linux
	./linux-generator
