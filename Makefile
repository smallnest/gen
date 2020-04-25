WORKDIR=`pwd`
export PACKR2_EXECUTABLE := $(shell command -v packr2  2> /dev/null)


default: build


check_prereq:
ifndef PACKR2_EXECUTABLE
	go get -u github.com/gobuffalo/packr/v2/packr2
endif
	$(warning "found packr2")


install:
	go get github.com/smallnest/gen

vet:
	go vet .

tools:
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u honnef.co/go/tools/cmd/gosimple
	go get -u honnef.co/go/tools/cmd/unused
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/fzipp/gocyclo
	go get -u github.com/golang/lint/golint
	go get -u github.com/gobuffalo/packr/v2/packr2

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
	go list -f '{{ join .Deps  "\n"}}' . |grep "/" | grep -v "github.com/smallnest/gen"| grep "\." | sort |uniq

fmt:
	go fmt .

build: check_prereq
	packr2 build .

test:
	go test  -v .


generate_example: clean_example
	ls -latr ./example
	go run . \
		--sqltype=sqlite3 \
		--connstr "./example/sample.db" \
		--database main  \
		--json \
		--gorm \
		--guregu \
		--rest \
		--out ./example \
		--module github.com/alexj212/generated \
		--mod \
		--server \
		--makefile \
		--overwrite

	cd ./example && $(MAKE) example

run_example: generate_example
	ls -latr ./example/bin
	./example/bin/example


clean_example:
	rm -rf ./example/Makefile \
	 ./example/README.md \
	 ./example/api \
	 ./example/app \
	 ./example/bin \
	 ./example/dao \
	 ./example/docs \
	 ./example/go.mod \
	 ./example/go.sum \
	 ./example/model \
	 ./example/.gitignore

