.PHONY: build clean deploy

deps:
	dep ensure -v

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/create-wallet lambdas/create-wallet/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/add-data lambdas/add-data/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list-data lambdas/list-data/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get-data lambdas/get-data/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get-data-history lambdas/get-data-history/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list-shared-data lambdas/list-shared-data/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get-shared-data lambdas/get-shared-data/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/share-data lambdas/share-data/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

test:
	go clean -testcache
	go test -v ./...

deploy: clean deps build
	sls deploy --verbose
