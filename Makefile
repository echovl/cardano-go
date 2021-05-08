GOBUILD = go build
GOTEST = go test

adaw:
	$(GOBUILD) -o ./build/cardano-wallet main.go

install:
	@mv ./build/cardano-wallet /usr/bin/

test:
	go test ./...
