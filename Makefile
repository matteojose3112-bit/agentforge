BINARY := bin/agentforge

.PHONY: test vet build run clean

test:
	go test ./...

vet:
	go vet ./...

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/agentforge

run:
	go run ./cmd/agentforge -tool echo -input "hello from AgentForge"

clean:
	rm -rf bin
