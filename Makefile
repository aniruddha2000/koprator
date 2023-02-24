.PHONY: docker-dev
docker-dev:
	docker-compose up

.PHONY: build
build:
	go build -o ./bin/koprator ./main.go
	gofmt -d .

.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	gofmt -s -w .
	golangci-lint run
