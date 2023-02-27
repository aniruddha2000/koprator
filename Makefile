.PHONY: docker
docker:
	docker build -t aniruddhabasak/koprator:latest .

.PHONY: docker-push
docker-push:
	docker push aniruddhabasak/koprator:latest


.PHONY: docker-dev
docker-dev:
	docker-compose up

.PHONY: build
build:
	go build -o ./bin/koprator ./main.go
	gofmt -d .

.PHONY: run
run:
	go run main.go -kubeconfig=$(HOME)/.kube/config

.PHONY: helm
helm:
	helm install koprator chart/koprator

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	gofmt -s -w .
	golangci-lint run