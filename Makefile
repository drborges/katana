.PHONY: clean update format test

test:
	@go test ./... -v -run=$(grep)

format:
	@go fmt ./...

build:
	@go build ./...

clean:
	@go clean

update:
	@go get -u -f github.com/smartystreets/goconvey/convey

rm-local-branches:
	@git branch | grep -v master | xargs -I {} git branch -D {}

