.PHONY: clean update code-check test

test:
	@go test ./... -v -run=$(grep)

tdd:
	@fswatch -o ./*.go | xargs -n1 -I{} make

code-check:
	@go fmt *.go
	@go fmt -s -r '(a) -> a' -w *.go
	@go vet
	@go fix

build:
	@go build ./...

clean:
	@go clean

update:
	@go get -u -f github.com/smartystreets/goconvey/convey

rm-local-branches:
	@git branch | grep -v master | xargs -I {} git branch -D {}

