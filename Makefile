GOBUILD = go build
GOTEST = go test

cardano-wallet:
	$(GOBUILD) -o ./cli/build/cardano-wallet cli/main.go

install:
	@cp ./cli/build/cardano-wallet /usr/bin/

test:
	$(GOTEST) ./...

testcov:
	$(GOTEST) ./... -coverprofile coverage.out

opencov:
	go tool cover -html coverage.out
