GOBUILD = go build
GOTEST = go test

cardano-wallet:
	$(GOBUILD) -o ./build/cardano-wallet main.go

install:
	@mv ./build/cardano-wallet /usr/bin/

test:
	$(GOTEST) ./...

testcov:
	$(GOTEST) ./... -coverprofile coverage.out

opencov:
	go tool cover -html coverage.out
