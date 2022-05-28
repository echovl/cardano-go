GOBUILD = go build
GOTEST = go test

cwallet:
	$(GOBUILD) -o ./cli/build/cwallet cli/main.go

install:
	@cp ./cli/build/cwallet /usr/bin/

test:
	$(GOTEST) ./...

testcov:
	$(GOTEST) ./... -coverprofile coverage.out

opencov:
	go tool cover -html coverage.out
