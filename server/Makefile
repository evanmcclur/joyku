# Go specific commands
tidy:
	go mod tidy

goclean:
	go clean -cache
	go clean -modcache

templgenerate:
	templ generate pkg/components/*.templ

# Build specific commands
joyku:
	go build -o bin/joyku_cli cmd/cli/*.go
	go build -o bin/joyku_web cmd/web/*.go

clean:
	rm -r bin/*
