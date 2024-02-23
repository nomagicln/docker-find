.PHONY: docker-find
docker-find:
	mkdir -p bin
	go build -o bin/docker-find ./cmd/find