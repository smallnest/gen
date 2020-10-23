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


install: check_prereq ## go install binary info $GOPATH/bin
	packr2 install github.com/smallnest/gen

vet: ## run go vet on the project
	go vet .

tools: ## install dependent tools
	go get -u github.com/gogo/protobuf
	go get -u github.com/gogo/protobuf/proto
	go get -u github.com/gogo/protobuf/jsonpb
	go get -u github.com/gogo/protobuf/protoc-gen-gogo
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u honnef.co/go/tools
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/fzipp/gocyclo
	go get -u golang.org/x/lint/golint
	go get -u github.com/gobuffalo/packr/v2/packr2

lint: ## run golint on the project
	golint ./...

staticcheck: ## run staticcheck on the project
	staticcheck -ignore "$(shell cat .checkignore)" .

unused:
	unused .

gocyclo: ## run gocyclo on the project
	@ gocyclo -over 20 $(shell find . -name "*.go" |egrep -v "pb\.go|_test\.go")

check: staticcheck gocyclo ## run code checks on the project

doc: ## run godoc
	godoc -http=:6060

deps:## analyze project deps
	go list -f '{{ join .Deps  "\n"}}' . |grep "/" | grep -v "github.com/smallnest/gen"| grep "\." | sort |uniq

fmt: ## run fmt on the project
	## go fmt .
	gofmt -s -d -w -l .

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
		--mapping=../template/mapping.json \
		--json \
		--db \
      	--api=apis \
      	--dao=daos \
      	--model=models \
		--generate-dao \
		--generate-proj \
		--protobuf \
		--gorm \
		--guregu \
		--rest \
		--mod \
		--server \
		--makefile \
		--run-gofmt \
		--copy-templates

test_exec: clean_example ## test example using sqlite db in ./examples
	ls -latr ./example
	cd ./custom && go run .. \
		--sqltype=sqlite3 \
		--connstr "../example/sample.db" \
		--database main \
		--module github.com/alexj212/generated \
		--context=./custom.json \
		--verbose \
		--overwrite \
		--out ./ \
		--exec=./sample.gen


build_example: generate_example ## generate and build example
	cd ./example && $(MAKE) example

run_example: example ## run example project server
	cd ./example && ./bin/example


clean_example: ## remove generated example code
	mkdir -p ./example
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
	 ./example/.gitignore \
	 ./tests






run_dbmeta: ## generate example project code from sqlite db in ./examples
	go run github.com/smallnest/gen/_test/dbmeta \
		--sqltype=sqlite3 \
		--connstr "./example/sample.db" \
		--database main

 ## --table employees_2


test: clean_example test_mysql test_postgres test_sqlite3 test_mssql  ## test mysql, mssql, postgres and sqlite3 code generation

test_mysql: ## test sqlite3 code generation
	./test.sh mysql  gen_sqlx && cd ./tests/mysql_sqlx && make example
	./test.sh mysql  gen_gorm && cd ./tests/mysql_gorm && make example
test_postgres: ## test postgres code generation
	./test.sh postgres  gen_sqlx && cd ./tests/postgres_sqlx && make example
	./test.sh postgres  gen_gorm && cd ./tests/postgres_gorm && make example
test_mssql: ## test mssql code generation
	./test.sh mssql  gen_sqlx && cd ./tests/mssql_sqlx && make example
	./test.sh mssql  gen_gorm && cd ./tests/mssql_gorm && make example
test_sqlite3: ## test sqlite3 code generation
	./test.sh sqlite3  gen_sqlx && cd ./tests/sqlite3_sqlx && make example
	./test.sh sqlite3  gen_gorm && cd ./tests/sqlite3_gorm && make example




set_release: ## populate release info
	./release.sh

gen_readme: set_release ## generate readme file
	go run github.com/smallnest/gen/readme \
		--sqltype=sqlite3 \
		--connstr "./example/sample.db" \
		--database main \
		--table invoices


release: gen_readme fmt gen install example  ## prepare release
	$(info ************  Release completed)

git_sync: ## git sync upstream
	git fetch upstream
	git checkout master
	git merge upstream/maste