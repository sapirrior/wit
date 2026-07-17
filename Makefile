TARGET = wit

all: build

build:
	CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" -o $(TARGET) cmd/wit/main.go

clean:
	rm -f $(TARGET)

.PHONY: all build clean
