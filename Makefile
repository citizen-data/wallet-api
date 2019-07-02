.PHONY: build clean deploy

build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/create-wallet lambdas/create-wallet/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/add-data lambdas/add-data/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list-data lambdas/list-data/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

test:
	go test -v ./...

deploy: clean build
	sls deploy --verbose
