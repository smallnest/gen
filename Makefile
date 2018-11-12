WORKDIR=`pwd`

default: build

install:
	go get github.com/nqsang90/gen

vet:
	go vet .

tools:
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u honnef.co/go/tools/cmd/gosimple
	go get -u honnef.co/go/tools/cmd/unused
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/fzipp/gocyclo
	go get -u github.com/golang/lint/golint

lint:
	golint ./...

staticcheck:
	staticcheck -ignore "$(shell cat .checkignore)" .

gosimple:
	# gosimple -ignore "$(shell cat .gosimpleignore)" .
	gosimple .

unused:
	unused .

gocyclo:
	@ gocyclo -over 20 $(shell find . -name "*.go" |egrep -v "pb\.go|_test\.go")

check: staticcheck gosimple unused gocyclo

doc:
	godoc -http=:6060

deps:
	go list -f '{{ join .Deps  "\n"}}' . |grep "/" | grep -v "github.com/nqsang90/gen"| grep "\." | sort |uniq

fmt:
	go fmt .

build:
	go build .

test:
	go test  -v .
