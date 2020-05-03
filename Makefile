WORKDIR=`pwd`
export PACKR2_EXECUTABLE := $(shell command -v packr2  2> /dev/null)

####################################################################################################################
##
## help for each task - https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
##
####################################################################################################################
.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

check_prereq: ## check pre requisites exist
ifndef PACKR2_EXECUTABLE
	go get -u github.com/gobuffalo/packr/v2/packr2
endif
	$(warning "found packr2")


install: ## go install binary info $GOPATH/bin
	packr2 install github.com/smallnest/gen

vet: ## run go vet on the project
	go vet .

tools: ## install dependent tools
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u honnef.co/go/tools/cmd/gosimple
	go get -u honnef.co/go/tools/cmd/unused
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/fzipp/gocyclo
	go get -u github.com/golang/lint/golint
	go get -u github.com/gobuffalo/packr/v2/packr2

lint: ## run golint on the project
	golint ./...

staticcheck: ## run staticcheck on the project
	staticcheck -ignore "$(shell cat .checkignore)" .

gosimple: ## run gosimple on the project
	# gosimple -ignore "$(shell cat .gosimpleignore)" .
	gosimple .

unused:
	unused .

gocyclo: ## run gocyclo on the project
	@ gocyclo -over 20 $(shell find . -name "*.go" |egrep -v "pb\.go|_test\.go")

check: staticcheck gosimple unused gocyclo ## run code checks on the project

doc: ## run godoc
	godoc -http=:6060

deps:## analyze project deps
	go list -f '{{ join .Deps  "\n"}}' . |grep "/" | grep -v "github.com/smallnest/gen"| grep "\." | sort |uniq

fmt: ## run fmt on the project
	go fmt .

build: check_prereq ## build gen binary
	packr2 build .

gen: build ## build gen binary

test: ## run go test on the project
	go test  -v .

example: generate_example ## generate example

generate_example: clean_example ## generate example project code from sqlite db in ./examples
	ls -latr ./example
	cd ./example && go run .. \
		--sqltype=sqlite3 \
		--connstr "./sample.db" \
		--database main \
		--module github.com/alexj212/generated \
		--verbose \
		--overwrite \
		--out ./ \
		--templateDir=../template \
		--json \
		--db \
		--protobuf \
		--gorm \
		--guregu \
		--rest \
		--mod \
		--server \
		--makefile \
		--copy-templates

test_exec: clean_example ## test example using sqlite db in ./examples
	ls -latr ./example
	cd ./custom && go run .. \
		--sqltype=sqlite3 \
		--connstr "../example/sample.db" \
		--database main \
		--module github.com/alexj212/generated \
		--verbose \
		--overwrite \
		--out ./ \
		--exec=./sample.gen


build_example: generate_example ## generate and build example
	cd ./example && $(MAKE) example

run_example: example ## run example project server
	cd ./example && ./bin/example


clean_example: ## remove generated example code
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

