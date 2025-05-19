# Makefile for building copyspace binaries for multiple platforms (static builds)

APP_NAME = copyspace

all: mac-arm64 mac-amd64 linux-amd64 linux-386 windows-amd64 windows-386

mac-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(APP_NAME)-mac-arm64 main.go

mac-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-mac-amd64 main.go

linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux-amd64 main.go

linux-386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o $(APP_NAME)-linux-386 main.go

windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(APP_NAME)-windows-amd64.exe main.go

windows-386:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o $(APP_NAME)-windows-386.exe main.go

clean:
	rm -f $(APP_NAME)-mac-arm64 $(APP_NAME)-mac-amd64 \
	      $(APP_NAME)-linux-amd64 $(APP_NAME)-linux-386 \
	      $(APP_NAME)-windows-amd64.exe $(APP_NAME)-windows-386.exe

.PHONY: all mac-arm64 mac-amd64 linux-amd64 linux-386 windows-amd64 windows-386 clean