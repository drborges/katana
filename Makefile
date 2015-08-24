.PHONY: all

gigo_path := $(GOPATH)/src/github.com/LyricalSecurity/gigo

all: setup update test
setup:
	@[[ ! -d $(gigo_path) ]] && git clone https://github.com/drborges/gigo $(gigo_path) || true
	go get github.com/LyricalSecurity/gigo/...
	go install github.com/LyricalSecurity/gigo

test:
	goapp test ./... -v -run=$(grep)

build:
	goapp build ./...

update:
	GIGO_GO=goapp gigo install -r requirements.txt

delete-branches:
	git branch | grep -v master | xargs -I {} git branch -D {}

