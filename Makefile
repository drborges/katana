.PHONY: clean update code-check test

test:
	@go test ./... -v -run=$(grep)

tdd:
	@fswatch -o ./*.go | xargs -n1 -I{} make

code-check:
	@gofmt -s -r '(a) -> a' -w *.go
	@go fmt
	@go vet
	@go fix

build:
	@go build ./...

clean:
	@go clean

update:
	@go get github.com/smartystreets/goconvey/convey
	@go get github.com/smartystreets/assertions

rm-local-branches:
	@git branch | grep -v master | xargs -I {} git branch -D {}

